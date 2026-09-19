package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

type RevocationService struct {
	Auth *UnitAuthService
}

func NewRevocationService() *RevocationService {
	return &RevocationService{Auth: NewUnitAuthService()}
}

// OpenRevocationCycle creates a new cycle targeting one membership.
// Only verified members of the same unit may initiate.
func (s *RevocationService) OpenRevocationCycle(
	unitID uuid.UUID,
	targetMembershipID uuid.UUID,
	initiatedByMembershipID uuid.UUID,
	cycleType string,
	reason string,
) (*models.RevocationCycle, error) {
	if cycleType != "regular_admin" && cycleType != "head_admin" {
		return nil, errors.New("cycleType must be regular_admin or head_admin")
	}

	var target models.UnitMembership
	if err := config.DB.First(&target, "id = ?", targetMembershipID).Error; err != nil {
		return nil, errors.New("target membership not found")
	}
	if target.UnitID != unitID {
		return nil, errors.New("target membership is not in this unit")
	}

	var initiator models.UnitMembership
	if err := config.DB.First(&initiator, "id = ?", initiatedByMembershipID).Error; err != nil {
		return nil, errors.New("initiator membership not found")
	}
	if initiator.UnitID != unitID || initiator.Status != models.MembershipActive || initiator.VerifiedAt == nil {
		return nil, errors.New("initiator must be a verified member of this unit")
	}

	var eligible int64
	config.DB.Model(&models.UnitMembership{}).
		Where("unit_id = ? AND status = ? AND verified_at IS NOT NULL", unitID, models.MembershipActive).
		Count(&eligible)

	if eligible == 0 {
		return nil, errors.New("no eligible verified members in unit")
	}

	policy, version, err := s.Auth.GetPolicy(unitID)
	if err != nil {
		return nil, err
	}

	required := policy.RegularAdminRemovalVotes
	if cycleType == "head_admin" {
		required = int(float64(eligible)*policy.HeadAdminRemovalMajority) + 1
	}

	quorum := int(float64(eligible) * policy.QuorumPercent)
	if quorum < 1 {
		quorum = 1
	}

	now := time.Now().UTC()
	cycle := models.RevocationCycle{
		UnitID:                  unitID,
		TargetMembershipID:      targetMembershipID,
		TargetRole:              target.Role,
		TargetMembershipStatus:  target.Status,
		CycleType:               cycleType,
		Status:                  "open",
		EligibleMemberCount:     int(eligible),
		EligibilitySnapshotAt:   now,
		QuorumRequired:          quorum,
		RequiredVotes:           required,
		InitiatedByMembershipID: initiatedByMembershipID,
		Reason:                  reason,
		UnitAuthPolicyVersion:   version,
		OpenedAt:                now,
	}

	if err := config.DB.Create(&cycle).Error; err != nil {
		return nil, err
	}
	return &cycle, nil
}

// CastVote records one member vote or one head-admin approval for a cycle.
func (s *RevocationService) CastVote(
	cycleID uuid.UUID,
	actorMembershipID uuid.UUID,
	actorUserID uuid.UUID,
	action string,
	choice string,
	reason string,
) error {
	if action != "member_vote" && action != "head_admin_approval" {
		return errors.New("action must be member_vote or head_admin_approval")
	}

	var cycle models.RevocationCycle
	if err := config.DB.First(&cycle, "id = ?", cycleID).Error; err != nil {
		return err
	}
	if cycle.Status != "open" {
		return errors.New("cycle is not open")
	}

	var actor models.UnitMembership
	if err := config.DB.First(&actor, "id = ?", actorMembershipID).Error; err != nil {
		return errors.New("actor membership not found")
	}
	if actor.UnitID != cycle.UnitID || actor.Status != models.MembershipActive || actor.VerifiedAt == nil {
		return errors.New("actor is not a verified member of the cycle unit")
	}

	var target models.UnitMembership
	if err := config.DB.First(&target, "id = ?", cycle.TargetMembershipID).Error; err != nil {
		return errors.New("target membership not found")
	}

	if action == "head_admin_approval" && !actor.IsHeadAdmin {
		return errors.New("only the head admin may record a head-admin approval")
	}

	vote := models.RevocationVote{
		CycleID:             cycleID,
		ActorMembershipID:   actorMembershipID,
		ActorUserID:         actorUserID,
		TargetMembershipID:  cycle.TargetMembershipID,
		TargetUserID:        target.UserID,
		Action:              action,
		Choice:              choice,
		Reason:              reason,
		EligibilityStatus:   "verified",
		EligibilitySnapshot: "active+verified",
		ActorRoleAtVote:     actor.Role,
		TargetRoleAtVote:    cycle.TargetRole,
		CreatedAt:           time.Now().UTC(),
	}

	if err := config.DB.Create(&vote).Error; err != nil {
		return err
	}

	return s.recomputeTally(cycleID)
}

// recomputeTally refreshes cycle counters from immutable vote rows.
func (s *RevocationService) recomputeTally(cycleID uuid.UUID) error {
	var cycle models.RevocationCycle
	if err := config.DB.First(&cycle, "id = ?", cycleID).Error; err != nil {
		return err
	}

	type row struct {
		Choice string
		Action string
		Count  int
	}
	var rows []row
	if err := config.DB.
		Model(&models.RevocationVote{}).
		Select("choice, action, COUNT(*) as count").
		Where("cycle_id = ?", cycleID).
		Group("choice, action").
		Scan(&rows).Error; err != nil {
		return err
	}

	forVotes := 0
	againstVotes := 0
	abstainVotes := 0
	total := 0
	approved := false

	for _, r := range rows {
		if r.Action == "head_admin_approval" {
			if r.Choice == "approve" {
				approved = true
			}
			continue
		}
		switch r.Choice {
		case "support":
			forVotes += r.Count
		case "oppose":
			againstVotes += r.Count
		case "abstain":
			abstainVotes += r.Count
		}
		total += r.Count
	}

	updates := map[string]interface{}{
		"for_votes":                 forVotes,
		"against_votes":             againstVotes,
		"abstain_votes":             abstainVotes,
		"total_votes":               total,
		"distinct_eligible_voters":  total,
		"head_admin_approved":       approved,
	}

	return config.DB.Model(&cycle).Updates(updates).Error
}

// CloseRevocationCycle finalizes the cycle, applying the removal if thresholds pass.
func (s *RevocationService) CloseRevocationCycle(cycleID uuid.UUID, headAdminApprovalMemberID *uuid.UUID) error {
	var cycle models.RevocationCycle
	if err := config.DB.First(&cycle, "id = ?", cycleID).Error; err != nil {
		return err
	}
	if cycle.Status != "open" {
		return errors.New("cycle is not open")
	}

	if err := s.recomputeTally(cycleID); err != nil {
		return err
	}
	if err := config.DB.First(&cycle, "id = ?", cycleID).Error; err != nil {
		return err
	}

	now := time.Now().UTC()
	policy, _, err := s.Auth.GetPolicy(cycle.UnitID)
	if err != nil {
		return err
	}

	quorumMet := cycle.TotalVotes >= cycle.QuorumRequired

	approved := false
	if cycle.CycleType == "regular_admin" {
		approved = quorumMet && cycle.ForVotes >= cycle.RequiredVotes && cycle.HeadAdminApproved
	} else {
		approved = quorumMet && cycle.ForVotes >= cycle.RequiredVotes
	}

	if !approved {
		return config.DB.Model(&cycle).Updates(map[string]interface{}{
			"status":    "rejected",
			"closed_at": now,
		}).Error
	}

	cooling := now.AddDate(0, policy.CoolingOffMonths, 0)

	if err := config.DB.Model(&models.UnitMembership{}).
		Where("id = ?", cycle.TargetMembershipID).
		Updates(map[string]interface{}{
			"status":            models.MembershipRevoked,
			"revoked_at":        now,
			"cooling_off_until": cooling,
		}).Error; err != nil {
		return err
	}

	return config.DB.Model(&cycle).Updates(map[string]interface{}{
		"status":            "completed",
		"closed_at":         now,
		"completed_at":      now,
		"cooling_off_until": cooling,
	}).Error
}