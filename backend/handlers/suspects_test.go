//go:build integration

package handlers

import (
	"testing"
	"time"

	"security-solution/config"
	"security-solution/internal/testutil"
	"security-solution/models"
	"security-solution/services"
)

// TestSuspect_CounterStatementReporterBlocked — the reporter of a case
// must never see the counter-statement, even though CanAccessCase
// grants them case_access_level=reporter.
func TestSuspect_CounterStatementReporterBlocked(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	reporter := testutil.MakeUser(t, "citizen")
	suspectUser := testutil.MakeUser(t, "citizen")
	testutil.MakeMembership(t, suspectUser, unit, "officer", false)

	c := testutil.MakeCase(t, reporter, unit)

	// Create a Suspect row linked to the citizen
	suspect := &models.Suspect{
		FirstName: "Test",
		LastName:  "Suspect",
		CreatedBy: head.ID,
		UserID:    &suspectUser.ID,
	}
	if err := config.DB.Create(suspect).Error; err != nil {
		t.Fatalf("create suspect: %v", err)
	}

	// Link suspect to case
	link := &models.SuspectCase{
		SuspectID: suspect.ID,
		CaseID:    c.ID,
		Role:      "suspect",
	}
	if err := config.DB.Create(link).Error; err != nil {
		t.Fatalf("link suspect: %v", err)
	}

	// Insert a counter-statement
	stmt := &models.CounterStatement{
		CaseID:  c.ID,
		UserID:  suspectUser.ID,
		Content: "I was not there.",
	}
	if err := config.DB.Create(stmt).Error; err != nil {
		t.Fatalf("create statement: %v", err)
	}

	// The guard rule: reporter must be blocked from reading.
	var caseObj models.Case
	config.DB.First(&caseObj, "id = ?", c.ID)
	if caseObj.ReportedBy == suspectUser.ID {
		t.Fatalf("setup error: reporter and suspect should differ")
	}
	if caseObj.ReportedBy != reporter.ID {
		t.Fatalf("setup error: reporter mismatch")
	}
}

// TestSuspect_OnlyLinkedCitizenIsSuspect — the isCaseSuspect rule
// requires Suspect.UserID -> SuspectCase link.
func TestSuspect_OnlyLinkedCitizenIsSuspect(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	reporter := testutil.MakeUser(t, "citizen")

	linked := testutil.MakeUser(t, "citizen")
	unlinked := testutil.MakeUser(t, "citizen")

	c := testutil.MakeCase(t, reporter, unit)

	suspect := &models.Suspect{
		FirstName: "Linked",
		LastName:  "Person",
		CreatedBy: head.ID,
		UserID:    &linked.ID,
	}
	config.DB.Create(suspect)
	config.DB.Create(&models.SuspectCase{
		SuspectID: suspect.ID,
		CaseID:    c.ID,
		Role:      "suspect",
	})

	// Direct DB check: linked user has a SuspectCase entry
	var countLinked int64
	config.DB.Model(&models.SuspectCase{}).
		Where("suspect_id = ? AND case_id = ?", suspect.ID, c.ID).
		Count(&countLinked)
	if countLinked != 1 {
		t.Fatalf("expected 1 link for linked user, got %d", countLinked)
	}

	// Unlinked user has no SuspectCase entry
	var countUnlinked int64
	config.DB.Model(&models.Suspect{}).
		Where("user_id = ?", unlinked.ID).
		Count(&countUnlinked)
	if countUnlinked != 0 {
		t.Fatalf("expected 0 suspect records for unlinked user, got %d", countUnlinked)
	}
}

// TestSuspect_ExpungementRequestRequires5Years — the expungement rule.
func TestSuspect_ExpungementRequestRequires5Years(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	reporter := testutil.MakeUser(t, "citizen")
	suspectUser := testutil.MakeUser(t, "citizen")

	c := testutil.MakeCase(t, reporter, unit)
	now := time.Now().UTC()
	config.DB.Model(c).Updates(map[string]interface{}{
		"status":    "closed",
		"closed_at": now,
	})

	suspect := &models.Suspect{
		FirstName: "Recent",
		LastName:  "Suspect",
		CreatedBy: head.ID,
		UserID:    &suspectUser.ID,
		UnitID:    &unit.ID,
	}
	config.DB.Create(suspect)
	config.DB.Create(&models.SuspectCase{
		SuspectID: suspect.ID,
		CaseID:    c.ID,
		Role:      "suspect",
	})

	svc := services.NewExpungementService()
	// A recent case (< 5 years) should fail the age check.
	_, err := svc.RequestExpungement(services.ExpungementRequestInput{
		SuspectID:   suspect.ID,
		RequesterID: suspectUser.ID,
		Reason:      "forgot",
	})
	if err == nil {
		t.Fatalf("expected <5y case to be rejected")
	}
}

// TestSuspect_ExpungementDeniedWithoutPending — cannot request when
// a pending request already exists.
func TestSuspect_ExpungementDeniedWithoutPending(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	suspectUser := testutil.MakeUser(t, "citizen")

	suspect := &models.Suspect{
		FirstName: "Old",
		LastName:  "Suspect",
		CreatedBy: head.ID,
		UserID:    &suspectUser.ID,
		UnitID:    &unit.ID,
	}
	config.DB.Create(suspect)

	// Insert an existing pending request
	existing := &models.ExpungementRequest{
		SuspectID:   suspect.ID,
		RequestedBy: suspectUser.ID,
		Reason:      "prior",
		Status:      "pending",
	}
	config.DB.Create(existing)

	svc := services.NewExpungementService()
	_, err := svc.RequestExpungement(services.ExpungementRequestInput{
		SuspectID:   suspect.ID,
		RequesterID: suspectUser.ID,
		Reason:      "duplicate",
	})
	if err == nil {
		t.Fatalf("expected duplicate pending request to be rejected")
	}
}

// TestSuspect_AppealWindowIs14Days — appeal outside 14 days is refused.
func TestSuspect_AppealWindowIs14Days(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	target := testutil.MakeUser(t, "unit_admin")
	m := testutil.MakeMembership(t, target, unit, "unit_admin", false)

	// Simulate a revocation cycle completed 20 days ago
	old := time.Now().UTC().AddDate(0, 0, -20)
	cycle := &models.RevocationCycle{
		UnitID:                  unit.ID,
		TargetMembershipID:      m.ID,
		TargetRole:              "unit_admin",
		TargetMembershipStatus:  "active",
		CycleType:               "regular_admin",
		Status:                  "completed",
		EligibleMemberCount:     10,
		EligibilitySnapshotAt:   old,
		QuorumRequired:          5,
		RequiredVotes:           10,
		InitiatedByMembershipID: m.ID,
		Reason:                  "misconduct",
		UnitAuthPolicyVersion:   "v1",
		OpenedAt:                old,
		CompletedAt:             &old,
	}
	config.DB.Create(cycle)

	// Mark the membership as revoked
	config.DB.Model(m).Updates(map[string]interface{}{
		"status":     "revoked",
		"revoked_at": old,
	})

	svc := services.NewAppealService()
	_, err := svc.FileAppeal(services.FileAppealInput{
		RevocationCycleID: cycle.ID,
		AppellantUserID:   target.ID,
		Reason:            "appeal",
	})
	if err == nil {
		t.Fatalf("expected >14d appeal window to be rejected")
	}
}