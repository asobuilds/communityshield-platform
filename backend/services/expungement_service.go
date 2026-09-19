package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

type ExpungementService struct{}

func NewExpungementService() *ExpungementService {
	return &ExpungementService{}
}

// ExpungementRequestInput is the citizen's request payload.
type ExpungementRequestInput struct {
	SuspectID   uuid.UUID
	RequesterID uuid.UUID
	Reason      string
}

const denialCooldownYears = 1
const eligibilityYears = 5

// RequestExpungement validates eligibility and creates a pending request.
// Guards run in order; each returns a distinct message.
func (s *ExpungementService) RequestExpungement(req ExpungementRequestInput) (*models.ExpungementRequest, error) {
	var suspect models.Suspect
	if err := config.DB.First(&suspect, "id = ?", req.SuspectID).Error; err != nil {
		return nil, errors.New("suspect not found")
	}

	// (a) only the citizen linked to this suspect may request expungement
	if suspect.UserID == nil || *suspect.UserID != req.RequesterID {
		return nil, errors.New("only the citizen linked to this suspect may request expungement")
	}

	// (b) no pending request may exist for this suspect
	var pending int64
	if err := config.DB.Model(&models.ExpungementRequest{}).
		Where("suspect_id = ? AND status = ?", req.SuspectID, "pending").
		Count(&pending).Error; err == nil && pending > 0 {
		return nil, errors.New("a pending expungement request already exists for this suspect")
	}

	// (c) no denial within the last 12 months
	cutoff := time.Now().UTC().AddDate(0, -denialCooldownYears, 0)
	var recentDenied int64
	if err := config.DB.Model(&models.ExpungementRequest{}).
		Where("suspect_id = ? AND status = ? AND reviewed_at > ?", req.SuspectID, "denied", cutoff).
		Count(&recentDenied).Error; err == nil && recentDenied > 0 {
		return nil, errors.New("expungement was denied within the last 12 months; please wait before requesting again")
	}

	// (d) every case linked to this suspect must be closed
	var openCount int64
	if err := config.DB.Model(&models.SuspectCase{}).
		Joins("JOIN cases ON cases.id = suspect_cases.case_id").
		Where("suspect_cases.suspect_id = ? AND cases.status != ?", req.SuspectID, "closed").
		Count(&openCount).Error; err == nil && openCount > 0 {
		return nil, errors.New("cannot expunge while a linked case is still open")
	}

	// (e) at least 5 years since the most recent case closure
	var mostRecent models.Case
	if err := config.DB.Model(&models.Case{}).
		Joins("JOIN suspect_cases ON suspect_cases.case_id = cases.id").
		Where("suspect_cases.suspect_id = ? AND cases.closed_at IS NOT NULL", req.SuspectID).
		Order("cases.closed_at DESC").
		First(&mostRecent).Error; err != nil {
		return nil, errors.New("no closed cases found; expungement requires at least 5 years since the most recent closed case")
	}
	lastClose := mostRecent.ClosedAt
	if lastClose == nil || time.Since(*lastClose) < time.Duration(eligibilityYears)*365*24*time.Hour {
		return nil, errors.New("expungement requires at least 5 years since the most recent closed case")
	}

	reqRecord := models.ExpungementRequest{
		SuspectID:   req.SuspectID,
		RequestedBy: req.RequesterID,
		RequestedAt: time.Now().UTC(),
		Reason:      req.Reason,
		Status:      "pending",
	}
	if err := config.DB.Create(&reqRecord).Error; err != nil {
		return nil, err
	}
	return &reqRecord, nil
}

// DecideExpungement approves or denies a pending request. Only a head admin of
// the suspect's unit (or a super admin) may decide.
func (s *ExpungementService) DecideExpungement(requestID, reviewerID uuid.UUID, decision, reason string) (*models.ExpungementRequest, error) {
	var reqRecord models.ExpungementRequest
	if err := config.DB.First(&reqRecord, "id = ?", requestID).Error; err != nil {
		return nil, errors.New("expungement request not found")
	}
	if reqRecord.Status != "pending" {
		return nil, errors.New("only a pending request can be decided")
	}

	var suspect models.Suspect
	config.DB.First(&suspect, "id = ?", reqRecord.SuspectID)

	// Authorization: head admin of the suspect's unit, or super admin.
	// If the suspect has no unit (UnitID nil), only super admin may decide.
	var reviewer models.User
	if err := config.DB.First(&reviewer, "id = ?", reviewerID).Error; err != nil {
		return nil, errors.New("reviewer not found")
	}
	if reviewer.IsSuperAdmin || reviewer.Role == "super_admin" {
		// allowed
	} else if suspect.UnitID != nil && *suspect.UnitID != uuid.Nil {
		var membership models.UnitMembership
		if err := config.DB.
			Where("unit_id = ? AND user_id = ? AND status = ?", *suspect.UnitID, reviewerID, models.MembershipActive).
			First(&membership).Error; err != nil || !membership.IsHeadAdmin {
			return nil, errors.New("only a head admin of the suspect's unit or a super admin may decide")
		}
	} else {
		return nil, errors.New("only a head admin of the suspect's unit or a super admin may decide")
	}

	now := time.Now().UTC()
	if decision != "approve" && decision != "deny" {
		return nil, errors.New("decision must be 'approve' or 'deny'")
	}
	reqRecord.Status = decision
	reqRecord.ReviewedBy = &reviewerID
	reqRecord.ReviewedAt = &now
	reqRecord.DecisionReason = reason

	if decision == "approve" {
		suspect.ExpungedAt = &now
		suspect.ExpungedBy = &reviewerID
		suspect.ExpungedReason = reason
		if err := config.DB.Save(&suspect).Error; err != nil {
			return nil, err
		}
	}
	if err := config.DB.Save(&reqRecord).Error; err != nil {
		return nil, err
	}
	return &reqRecord, nil
}

// ListRequests returns expungement requests for a given suspect.
func (s *ExpungementService) ListRequests(suspectID, requesterID uuid.UUID) ([]models.ExpungementRequest, error) {
	var suspect models.Suspect
	if err := config.DB.First(&suspect, "id = ?", suspectID).Error; err != nil {
		return nil, errors.New("suspect not found")
	}
	if suspect.UserID == nil || *suspect.UserID != requesterID {
		return nil, errors.New("only the linked citizen may view expungement requests for this suspect")
	}

	var reqs []models.ExpungementRequest
	if err := config.DB.Where("suspect_id = ?", suspectID).Order("created_at DESC").Find(&reqs).Error; err != nil {
		return nil, err
	}
	return reqs, nil
}
