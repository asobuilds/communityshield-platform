package services

import (
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

type ElectionService struct {
	Auth *UnitAuthService
}

func NewElectionService() *ElectionService {
	return &ElectionService{Auth: NewUnitAuthService()}
}

// eligibleVerifiedMembers returns active memberships with verified_at set.
func (s *ElectionService) eligibleVerifiedMembers(unitID uuid.UUID) ([]models.UnitMembership, error) {
	var members []models.UnitMembership
	err := config.DB.
		Where("unit_id = ? AND status = ? AND verified_at IS NOT NULL", unitID, models.MembershipActive).
		Find(&members).Error
	return members, err
}

// SeatCountForMembers picks the seat count for a given member count using the UnitAuth seat bands.
func SeatCountForMembers(policy UnitAuthPolicy, memberCount int) int {
	switch {
	case memberCount >= 12 && memberCount <= 20:
		return policy.SeatBands["12-20"]
	case memberCount >= 21 && memberCount <= 40:
		return policy.SeatBands["21-40"]
	case memberCount >= 41 && memberCount <= 70:
		return policy.SeatBands["41-70"]
	case memberCount >= 71 && memberCount <= 100:
		return policy.SeatBands["71-100"]
	}
	return 0
}

// OpenAdminElection creates a UnitAdminElection record in draft status.
func (s *ElectionService) OpenAdminElection(unitID uuid.UUID, createdBy uuid.UUID, rotationGroup string) (*models.UnitAdminElection, error) {
	policy, _, err := s.Auth.GetPolicy(unitID)
	if err != nil {
		return nil, err
	}

	members, err := s.eligibleVerifiedMembers(unitID)
	if err != nil {
		return nil, err
	}
	if len(members) < 12 {
		return nil, errors.New("unit must have at least 12 verified members before an election")
	}

	seatCount := SeatCountForMembers(policy, len(members))
	if seatCount == 0 {
		return nil, errors.New("member count does not fall into a supported seat band")
	}

	quorum := int(float64(len(members)) * policy.QuorumPercent)
	if quorum < 1 {
		quorum = 1
	}

	now := time.Now().UTC()
	termEnd := now.AddDate(0, policy.TermMonths, 0)
	votingEnds := now.AddDate(0, 0, 7)

	election := models.UnitAdminElection{
		UnitID:                unitID,
		ElectionType:          "admin",
		RotationGroup:         rotationGroup,
		SeatCount:             seatCount,
		MemberCountAtElection: len(members),
		EligibleVoterCount:    len(members),
		QuorumCount:           quorum,
		TermStart:             now,
		TermEnd:               termEnd,
		VotingStartsAt:        &now,
		VotingEndsAt:          &votingEnds,
		Status:                "open",
		CreatedBy:             createdBy,
	}

	if err := config.DB.Create(&election).Error; err != nil {
		return nil, err
	}
	return &election, nil
}

// CastAdminVote records a verified member's vote for an admin candidate.
func (s *ElectionService) CastAdminVote(electionID uuid.UUID, voterUserID uuid.UUID, candidateUserID uuid.UUID) error {
	var election models.UnitAdminElection
	if err := config.DB.First(&election, "id = ?", electionID).Error; err != nil {
		return err
	}
	if election.Status != "open" {
		return errors.New("election is not open")
	}
	if election.VotingEndsAt != nil && time.Now().UTC().After(*election.VotingEndsAt) {
		return errors.New("voting window has closed")
	}

	var voter models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ? AND verified_at IS NOT NULL",
			election.UnitID, voterUserID, models.MembershipActive).
		First(&voter).Error; err != nil {
		return errors.New("voter is not a verified member of this unit")
	}

	var candidate models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ? AND verified_at IS NOT NULL",
			election.UnitID, candidateUserID, models.MembershipActive).
		First(&candidate).Error; err != nil {
		return errors.New("candidate is not a verified member of this unit")
	}

	vote := models.AdminVote{
		ElectionID:  electionID,
		VoterID:     voterUserID,
		CandidateID: candidateUserID,
	}
	return config.DB.Create(&vote).Error
}

// CloseAdminElection tallies votes, seats top-N candidates, promotes them,
// and sets term/cooling-off fields on their memberships.
func (s *ElectionService) CloseAdminElection(electionID uuid.UUID) error {
	var election models.UnitAdminElection
	if err := config.DB.First(&election, "id = ?", electionID).Error; err != nil {
		return err
	}
	if election.Status != "open" {
		return errors.New("election is not open")
	}

	policy, _, err := s.Auth.GetPolicy(election.UnitID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	// Tie-break order (matches WardGuard's fairness rule):
	//   1. Higher vote count wins
	//   2. Longer unit membership wins (earlier joined_at)
	//   3. Older platform account wins (earlier user_created)
	//   4. Deterministic final fallback: candidate UUID ascending
	// The final UUID sort exists only to make results reproducible;
	// it does not favor any candidate.
	type tally struct {
		CandidateID uuid.UUID
		Count       int
		JoinedAt    time.Time
		UserCreated time.Time
	}
	var rows []tally
	if err := config.DB.
		Model(&models.AdminVote{}).
		Select(`admin_votes.candidate_id,
		        COUNT(*) AS count,
		        COALESCE(unit_memberships.created_at, NOW()) AS joined_at,
		        COALESCE(users.created_at, NOW()) AS user_created`).
		Joins("LEFT JOIN unit_memberships ON unit_memberships.user_id = admin_votes.candidate_id AND unit_memberships.unit_id = ?", election.UnitID).
		Joins("LEFT JOIN users ON users.id = admin_votes.candidate_id").
		Where("admin_votes.election_id = ?", electionID).
		Group("admin_votes.candidate_id, unit_memberships.created_at, users.created_at").
		Order("count DESC, joined_at ASC, user_created ASC, admin_votes.candidate_id ASC").
		Scan(&rows).Error; err != nil {
		return err
	}

	totalVotes := 0
	for _, r := range rows {
		totalVotes += r.Count
	}
	quorumMet := totalVotes >= election.QuorumCount

	// First failure: extend once by 7 days, stay open.
	if !quorumMet && !election.ExtendedOnce {
		newEnd := now.AddDate(0, 0, 7)
		return config.DB.Model(&election).Updates(map[string]interface{}{
			"extended_once":  true,
			"voting_ends_at": newEnd,
		}).Error
	}

	// Second failure (or already extended): pass with shortened term.
	termEnd := election.TermEnd
	lowTurnout := !quorumMet
	if lowTurnout {
		termEnd = election.TermStart.AddDate(0, 6, 0)
	}

	winnerCount := election.SeatCount
	if winnerCount > len(rows) {
		winnerCount = len(rows)
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i := 0; i < winnerCount; i++ {
		winner := rows[i]

		seat := models.UnitAdminSeat{
			UnitID:        election.UnitID,
			ElectionID:    election.ID,
			SeatNumber:    i + 1,
			RotationGroup: election.RotationGroup,
			MemberID:      &winner.CandidateID,
			TermStart:     election.TermStart,
			TermEnd:       termEnd,
			Status:        "active",
			ElectedAt:     now,
		}
		if err := tx.Create(&seat).Error; err != nil {
			tx.Rollback()
			return err
		}

		var membership models.UnitMembership
		if err := tx.
			Where("unit_id = ? AND user_id = ?", election.UnitID, winner.CandidateID).
			First(&membership).Error; err != nil {
			continue
		}

		termStart := election.TermStart
		membership.Role = models.UnitRoleAdmin
		membership.ElectedAt = &now
		membership.TermStartAt = &termStart
		membership.TermEndAt = &termEnd
		membership.ConsecutiveTerms = membership.ConsecutiveTerms + 1

		if membership.ConsecutiveTerms >= policy.ConsecutiveTermLimit {
			cooling := termEnd.AddDate(0, policy.CoolingOffMonths, 0)
			membership.CoolingOffUntil = &cooling
		}

		if err := tx.Save(&membership).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	finalStatus := "finalized"
	if lowTurnout {
		finalStatus = "finalized_low_turnout"
	}

	if err := tx.Model(&election).Updates(map[string]interface{}{
		"status":              finalStatus,
		"result_finalized_at": now,
		"quorum_met":          quorumMet,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// RunHeadAdminElection runs the head-admin vote among current admins and
// promotes the majority winner to IsHeadAdmin=true.
func (s *ElectionService) RunHeadAdminElection(unitID uuid.UUID, createdBy uuid.UUID) error {
	var admins []models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND status = ? AND role = ?",
			unitID, models.MembershipActive, models.UnitRoleAdmin).
		Find(&admins).Error; err != nil {
		return err
	}
	if len(admins) < 3 {
		return errors.New("at least 3 admins required to run a head-admin election")
	}

	election := models.UnitAdminElection{
		UnitID:        unitID,
		ElectionType:  "head_admin",
		RotationGroup: "head",
		SeatCount:     1,
		Status:        "open",
		TermStart:     time.Now().UTC(),
		TermEnd:       time.Now().UTC().AddDate(1, 0, 0),
		CreatedBy:     createdBy,
	}
	return config.DB.Create(&election).Error
}

// FinalizeHeadAdminElection tallies votes and sets the winner as head admin.
func (s *ElectionService) FinalizeHeadAdminElection(electionID uuid.UUID) error {
	var election models.UnitAdminElection
	if err := config.DB.First(&election, "id = ?", electionID).Error; err != nil {
		return err
	}
	if election.ElectionType != "head_admin" || election.Status != "open" {
		return errors.New("not an open head-admin election")
	}

	type tally struct {
		CandidateID uuid.UUID
		Count       int
	}
	var rows []tally
	if err := config.DB.
		Model(&models.HeadAdminVote{}).
		Select("candidate_id, COUNT(*) as count").
		Where("election_id = ?", electionID).
		Group("candidate_id").
		Order("count DESC").
		Scan(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return errors.New("no votes cast")
	}

	winner := rows[0]
	now := time.Now().UTC()

	tx := config.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.UnitMembership{}).
		Where("unit_id = ? AND is_head_admin = ?", election.UnitID, true).
		Update("is_head_admin", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.UnitMembership{}).
		Where("unit_id = ? AND user_id = ?", election.UnitID, winner.CandidateID).
		Updates(map[string]interface{}{
			"is_head_admin": true,
			"elected_at":    now,
			"term_start_at": election.TermStart,
			"term_end_at":   election.TermEnd,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	seat := models.UnitAdminSeat{
		UnitID:        election.UnitID,
		ElectionID:    election.ID,
		SeatNumber:    1,
		RotationGroup: "head",
		MemberID:      &winner.CandidateID,
		TermStart:     election.TermStart,
		TermEnd:       election.TermEnd,
		Status:        "active",
		ElectedAt:     now,
	}
	if err := tx.Create(&seat).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&election).Updates(map[string]interface{}{
		"status":              "finalized",
		"result_finalized_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// sortSeatsByRotationGroup is a helper for stagger assignment (service-internal).
func sortSeatsByRotationGroup(seats []models.UnitAdminSeat) {
	sort.SliceStable(seats, func(i, j int) bool {
		return seats[i].SeatNumber < seats[j].SeatNumber
	})
}

// FillVacancy seats the next-highest vote-getter from the given election
// into the seat identified by seatID. Uses the same tie-break order as
// CloseAdminElection. No-op if the seat is not vacant, the election is
// missing, or no eligible candidate remains.
func (s *ElectionService) FillVacancy(seatID uuid.UUID) error {
	var seat models.UnitAdminSeat
	if err := config.DB.First(&seat, "id = ?", seatID).Error; err != nil {
		return errors.New("seat not found")
	}
	if seat.Status != "vacant" {
		return errors.New("seat is not vacant")
	}

	var election models.UnitAdminElection
	if err := config.DB.First(&election, "id = ?", seat.ElectionID).Error; err != nil {
		return errors.New("election not found")
	}

	// Candidates who already hold a seat in this election
	var seatedIDs []uuid.UUID
	config.DB.Model(&models.UnitAdminSeat{}).
		Where("election_id = ? AND member_id IS NOT NULL", election.ID).
		Pluck("member_id", &seatedIDs)

	type tally struct {
		CandidateID uuid.UUID
		Count       int
		JoinedAt    time.Time
		UserCreated time.Time
	}
	query := config.DB.
		Model(&models.AdminVote{}).
		Select(`admin_votes.candidate_id,
		        COUNT(*) AS count,
		        COALESCE(unit_memberships.created_at, NOW()) AS joined_at,
		        COALESCE(users.created_at, NOW()) AS user_created`).
		Joins("LEFT JOIN unit_memberships ON unit_memberships.user_id = admin_votes.candidate_id AND unit_memberships.unit_id = ?", election.UnitID).
		Joins("LEFT JOIN users ON users.id = admin_votes.candidate_id").
		Where("admin_votes.election_id = ?", election.ID).
		Group("admin_votes.candidate_id, unit_memberships.created_at, users.created_at").
		Order("count DESC, joined_at ASC, user_created ASC, admin_votes.candidate_id ASC")

	if len(seatedIDs) > 0 {
		query = query.Where("admin_votes.candidate_id NOT IN ?", seatedIDs)
	}

	var rows []tally
	if err := query.Scan(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return errors.New("no eligible candidate for vacancy")
	}

	next := rows[0]
	now := time.Now().UTC()

	var membership models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ?", election.UnitID, next.CandidateID).
		First(&membership).Error; err != nil {
		return errors.New("candidate membership not found")
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	termEnd := seat.TermEnd
	candidateID := next.CandidateID

	if err := tx.Model(&seat).Updates(map[string]interface{}{
		"member_id":   candidateID,
		"status":      "active",
		"elected_at":  now,
		"vacated_at":  nil,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	termStart := seat.TermStart
	membership.Role = models.UnitRoleAdmin
	membership.ElectedAt = &now
	membership.TermStartAt = &termStart
	membership.TermEndAt = &termEnd

	if err := tx.Save(&membership).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
