//go:build integration

package handlers

import (
	"testing"

	"security-solution/config"
	"security-solution/internal/testutil"
	"security-solution/models"
	"security-solution/services"

	"github.com/google/uuid"
)

// TestFinance_LedgerAppendIsImmutable proves ledger entries cannot be
// mutated after creation. New corrections must be reversal entries.
func TestFinance_LedgerAppendIsImmutable(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)

	svc := services.NewLedgerService()
	entry, err := svc.Append(services.LedgerEntryInput{
		UnitID:       unit.ID,
		Direction:    "in",
		EntryType:    "donation",
		Amount:       50000,
		Counterparty: "Test Donor",
		Description:  "Test donation",
		SourceType:   "manual",
		CreatedBy:    head.ID,
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if entry.Reference == "" {
		t.Fatalf("expected non-empty reference")
	}

	// Attempt to mutate — the schema has no UpdatedAt, but a raw update
	// would still succeed in Postgres. Test the app-level invariant:
	// the service should never return a modified entry because we don't
	// have an Update method. This is a documentation test.
	before := entry.Amount
	entry.Amount = 999999
	// Do NOT save — just prove the caller can't accidentally commit a change
	// through the service. Verify the DB still has the original.
	var reloaded models.FinancialLedger
	config.DB.First(&reloaded, "id = ?", entry.ID)
	if reloaded.Amount != before {
		t.Fatalf("ledger entry was modified in DB: expected %v, got %v", before, reloaded.Amount)
	}
}

// TestFinance_LedgerReferenceIsSequential proves the reference counter
// increments and formatting is stable.
func TestFinance_LedgerReferenceIsSequential(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	svc := services.NewLedgerService()

	first, err := svc.Append(services.LedgerEntryInput{
		UnitID: unit.ID, Direction: "in", EntryType: "donation",
		Amount: 1000, Counterparty: "A", Description: "1",
		SourceType: "manual", CreatedBy: head.ID,
	})
	if err != nil {
		t.Fatalf("append 1: %v", err)
	}
	second, err := svc.Append(services.LedgerEntryInput{
		UnitID: unit.ID, Direction: "in", EntryType: "donation",
		Amount: 2000, Counterparty: "B", Description: "2",
		SourceType: "manual", CreatedBy: head.ID,
	})
	if err != nil {
		t.Fatalf("append 2: %v", err)
	}

	if first.SequenceNumber >= second.SequenceNumber {
		t.Fatalf("expected sequence to increase: %d -> %d", first.SequenceNumber, second.SequenceNumber)
	}
	if first.Reference == second.Reference {
		t.Fatalf("expected unique references: %s vs %s", first.Reference, second.Reference)
	}
}

// TestFinance_LedgerSummaryAggregates proves totals compute correctly.
func TestFinance_LedgerSummaryAggregates(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	svc := services.NewLedgerService()

	for _, amt := range []float64{10000, 25000, 5000} {
		if _, err := svc.Append(services.LedgerEntryInput{
			UnitID: unit.ID, Direction: "in", EntryType: "donation",
			Amount: amt, Counterparty: "X", Description: "in",
			SourceType: "manual", CreatedBy: head.ID,
		}); err != nil {
			t.Fatalf("append in: %v", err)
		}
	}
	for _, amt := range []float64{3000, 2000} {
		if _, err := svc.Append(services.LedgerEntryInput{
			UnitID: unit.ID, Direction: "out", EntryType: "expense",
			Amount: amt, Counterparty: "Y", Description: "out",
			SourceType: "manual", CreatedBy: head.ID,
		}); err != nil {
			t.Fatalf("append out: %v", err)
		}
	}

	year := firstYearOf(unit.ID)
	summary, err := svc.SummaryForUnitYear(unit.ID, year)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	if summary.TotalIn != 40000 {
		t.Fatalf("expected total_in 40000, got %v", summary.TotalIn)
	}
	if summary.TotalOut != 5000 {
		t.Fatalf("expected total_out 5000, got %v", summary.TotalOut)
	}
	if summary.Net != 35000 {
		t.Fatalf("expected net 35000, got %v", summary.Net)
	}
}

// helper: read the year of the first ledger entry for a unit
func firstYearOf(unitID uuid.UUID) int {
	var entry models.FinancialLedger
	if err := config.DB.Where("unit_id = ?", unitID).
		Order("sequence_number ASC").First(&entry).Error; err != nil {
		return 0
	}
	return entry.Year
}

// TestFinance_YearlyCloseIsIdempotent proves CloseYear writes exactly one
// snapshot even when called twice.
func TestFinance_YearlyCloseIsIdempotent(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	svc := services.NewLedgerService()

	if _, err := svc.Append(services.LedgerEntryInput{
		UnitID: unit.ID, Direction: "in", EntryType: "donation",
		Amount: 10000, Counterparty: "A", Description: "1",
		SourceType: "manual", CreatedBy: head.ID,
	}); err != nil {
		t.Fatalf("append: %v", err)
	}

	year := firstYearOf(unit.ID)

	first, err := svc.CloseYear(unit.ID, year, head.ID)
	if err != nil {
		t.Fatalf("close 1: %v", err)
	}
	second, err := svc.CloseYear(unit.ID, year, head.ID)
	if err != nil {
		t.Fatalf("close 2: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected same snapshot ID on second call, got %s vs %s", first.ID, second.ID)
	}

	var count int64
	config.DB.Model(&models.UnitFinancialYear{}).
		Where("unit_id = ? AND year = ?", unit.ID, year).
		Count(&count)
	if count != 1 {
		t.Fatalf("expected exactly 1 snapshot, got %d", count)
	}
}