//go:build integration

package handlers

import (
	"testing"
	"time"

	"security-solution/config"
	"security-solution/internal/testutil"
	"security-solution/models"
	"security-solution/services"

	"github.com/google/uuid"
)

// makeElection builds an open election with N verified members and M votes.
func makeElection(t *testing.T, unit *models.SecurityUnit, seatCount, quorum int, votesPerCandidate []int) *models.UnitAdminElection {
	t.Helper()

	now := time.Now().UTC()
	election := &models.UnitAdminElection{
		UnitID:             unit.ID,
		ElectionType:       "admin",
		RotationGroup:      "A",
		SeatCount:          seatCount,
		MemberCountAtElection: len(votesPerCandidate),
		EligibleVoterCount: len(votesPerCandidate),
		QuorumCount:        quorum,
		TermStart:          now,
		TermEnd:            now.AddDate(1, 0, 0),
		Status:             "open",
		CreatedBy:          uuid.New(),
	}
	if err := config.DB.Create(election).Error; err != nil {
		t.Fatalf("create election: %v", err)
	}

	for candidateIdx, voteCount := range votesPerCandidate {
		candidate := testutil.MakeUser(t, "citizen")
		testutil.MakeMembership(t, candidate, unit, "officer", false)

		for v := 0; v < voteCount; v++ {
			voter := testutil.MakeUser(t, "citizen")
			vote := &models.AdminVote{
				ElectionID:  election.ID,
				VoterID:     voter.ID,
				CandidateID: candidate.ID,
			}
			if err := config.DB.Create(vote).Error; err != nil {
				t.Fatalf("create vote: %v", err)
			}
			_ = candidateIdx
		}
	}

	return election
}

// TestElection_FirstQuorumMissExtends extends once and keeps the election open.
func TestElection_FirstQuorumMissExtends(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	election := makeElection(t, unit, 5, 100, []int{3, 2, 1})

	svc := services.NewElectionService()
	if err := svc.CloseAdminElection(election.ID); err != nil {
		t.Fatalf("close first time: %v", err)
	}

	var refreshed models.UnitAdminElection
	if err := config.DB.First(&refreshed, "id = ?", election.ID).Error; err != nil {
		t.Fatalf("reload election: %v", err)
	}
	if refreshed.Status != "open" {
		t.Fatalf("expected status open after first miss, got %s", refreshed.Status)
	}
	if !refreshed.ExtendedOnce {
		t.Fatalf("expected extended_once=true after first miss")
	}
	if refreshed.VotingEndsAt == nil || refreshed.VotingEndsAt.Before(time.Now().UTC()) {
		t.Fatalf("expected future voting_ends_at after extension")
	}
}

// TestElection_SecondQuorumMissFinalizesLowTurnout finalizes with 6-month term.
func TestElection_SecondQuorumMissFinalizesLowTurnout(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	election := makeElection(t, unit, 3, 100, []int{3, 2, 1})

	svc := services.NewElectionService()
	// First close: extends
	if err := svc.CloseAdminElection(election.ID); err != nil {
		t.Fatalf("first close: %v", err)
	}
	// Second close: finalizes with low turnout
	if err := svc.CloseAdminElection(election.ID); err != nil {
		t.Fatalf("second close: %v", err)
	}

	var refreshed models.UnitAdminElection
	config.DB.First(&refreshed, "id = ?", election.ID)
	if refreshed.Status != "finalized_low_turnout" {
		t.Fatalf("expected status finalized_low_turnout, got %s", refreshed.Status)
	}

	var seat models.UnitAdminSeat
	if err := config.DB.Where("election_id = ?", election.ID).First(&seat).Error; err != nil {
		t.Fatalf("expected at least one seat: %v", err)
	}
	// Term should be 6 months from TermStart, not 12
	expectedEnd := refreshed.TermStart.AddDate(0, 6, 0)
	if seat.TermEnd.Sub(expectedEnd).Abs() > time.Hour {
		t.Fatalf("expected 6-month term end %v, got %v", expectedEnd, seat.TermEnd)
	}
}

// TestElection_TieBreakOrdersBySeniority proves tie-break ordering.
func TestElection_TieBreakOrdersBySeniority(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)

	now := time.Now().UTC()
	election := &models.UnitAdminElection{
		UnitID:             unit.ID,
		ElectionType:       "admin",
		RotationGroup:      "A",
		SeatCount:          1,
		MemberCountAtElection: 2,
		EligibleVoterCount: 2,
		QuorumCount:        1,
		TermStart:          now,
		TermEnd:            now.AddDate(1, 0, 0),
		Status:             "open",
		CreatedBy:          uuid.New(),
	}
	config.DB.Create(election)

	// Candidate 1: newer member
	newer := testutil.MakeUser(t, "citizen")
	m1 := testutil.MakeMembership(t, newer, unit, "officer", false)
	// Bump joined_at forward
	config.DB.Model(m1).Update("created_at", now)

	// Candidate 2: older member
	older := testutil.MakeUser(t, "citizen")
	m2 := testutil.MakeMembership(t, older, unit, "officer", false)
	config.DB.Model(m2).Update("created_at", now.AddDate(0, -6, 0))

	// Both get 1 vote — tie
	voterA := testutil.MakeUser(t, "citizen")
	voterB := testutil.MakeUser(t, "citizen")
	config.DB.Create(&models.AdminVote{ElectionID: election.ID, VoterID: voterA.ID, CandidateID: newer.ID})
	config.DB.Create(&models.AdminVote{ElectionID: election.ID, VoterID: voterB.ID, CandidateID: older.ID})

	svc := services.NewElectionService()
	if err := svc.CloseAdminElection(election.ID); err != nil {
		t.Fatalf("close: %v", err)
	}

	var seat models.UnitAdminSeat
	if err := config.DB.Where("election_id = ?", election.ID).First(&seat).Error; err != nil {
		t.Fatalf("expected seat: %v", err)
	}
	if seat.MemberID == nil || *seat.MemberID != older.ID {
		t.Fatalf("expected older member %s to win tie-break, got %v", older.ID, seat.MemberID)
	}
}

// TestElection_VacancyFillSeatsNextCandidate proves auto-fill logic.
func TestElection_VacancyFillSeatsNextCandidate(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	election := makeElection(t, unit, 2, 1, []int{3, 2, 1})

	svc := services.NewElectionService()
	if err := svc.CloseAdminElection(election.ID); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Vacate seat 1
	var seat models.UnitAdminSeat
	config.DB.Where("election_id = ? AND seat_number = ?", election.ID, 1).First(&seat)
	config.DB.Model(&seat).Updates(map[string]interface{}{"status": "vacant", "vacated_at": time.Now().UTC()})

	if err := svc.FillVacancy(seat.ID); err != nil {
		t.Fatalf("fill vacancy: %v", err)
	}

	var refreshed models.UnitAdminSeat
	config.DB.First(&refreshed, "id = ?", seat.ID)
	if refreshed.Status != "active" {
		t.Fatalf("expected seat active, got %s", refreshed.Status)
	}
	if refreshed.MemberID == nil {
		t.Fatalf("expected member_id populated after fill")
	}
}
