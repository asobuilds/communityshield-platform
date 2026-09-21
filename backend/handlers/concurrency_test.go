//go:build integration

package handlers

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"security-solution/config"
	"security-solution/internal/testutil"
	"security-solution/models"
	"security-solution/services"
)

// TestFinancialLedgerConcurrentAppend proves three properties when N
// goroutines call LedgerService.Append concurrently against the same
// unit/year scope:
//
//   (A) No reference-sequencing race — every returned reference is unique.
//   (B) No duplicate references in the DB — COUNT(*) == COUNT(DISTINCT reference).
//   (C) No lost updates — every successful append has a committed row and
//       the sum of amounts matches the DB aggregate.
//
// The Append implementation generates references via an atomic counter
// table (ledger_sequences) updated inside the same transaction as the
// ledger insert, so no duplicate references should ever occur and no
// retries should be needed.
func TestFinancialLedgerConcurrentAppend(t *testing.T) {
	testutil.TruncateAll(t)
	defer testutil.TruncateAll(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	svc := services.NewLedgerService()

	const N = 50

	var (
		mu         sync.Mutex
		references []string
		amounts    []float64
		successes  int
		startCh    = make(chan struct{})
		wg         sync.WaitGroup
	)

	// Seed one entry so the counter baseline is non-zero,
	// making any off-by-one / duplicate collision easier to spot.
	if _, err := svc.Append(services.LedgerEntryInput{
		UnitID: unit.ID, Direction: "in", EntryType: "donation",
		Amount: 1, Counterparty: "seed", Description: "seed",
		SourceType: "manual", CreatedBy: head.ID,
	}); err != nil {
		t.Fatalf("seed append: %v", err)
	}

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-startCh // barrier — all goroutines contend at once

			select {
			case <-ctx.Done():
				return
			default:
			}

			entry, err := svc.Append(services.LedgerEntryInput{
				UnitID: unit.ID, Direction: "in", EntryType: "donation",
				Amount: float64(i + 1),
				Counterparty: fmt.Sprintf("donor-%d", i),
				Description: fmt.Sprintf("concurrent-%d", i),
				SourceType: "manual", CreatedBy: head.ID,
			})

			mu.Lock()
			if err != nil {
				t.Errorf("Append failed (no retry expected): %v", err)
			} else {
				successes++
				references = append(references, entry.Reference)
				amounts = append(amounts, entry.Amount)
			}
			mu.Unlock()
		}(i)
	}

	close(startCh) // release the barrier
	wg.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("timeout: goroutines deadlocked")
	}

	// (A) No duplicate references returned by the service.
	seen := make(map[string]bool, len(references))
	for _, ref := range references {
		if seen[ref] {
			t.Fatalf("(A) duplicate reference returned: %s", ref)
		}
		seen[ref] = true
	}

	// (B) DB-level uniqueness: COUNT(*) must equal COUNT(DISTINCT reference).
	var total, distinct int64
	if err := config.DB.Model(&models.FinancialLedger{}).
		Where("unit_id = ?", unit.ID).
		Count(&total).Error; err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if err := config.DB.Model(&models.FinancialLedger{}).
		Where("unit_id = ?", unit.ID).
		Select("COUNT(DISTINCT reference)").
		Scan(&distinct).Error; err != nil {
		t.Fatalf("count distinct: %v", err)
	}
	if total != distinct {
		t.Fatalf("(B) DB has %d rows but only %d distinct references", total, distinct)
	}

	// (C) No lost updates: every success has a committed row, and the
	// sum of amounts in the DB equals the seed (1) plus the sum of all
	// amounts actually returned by successful appends.
	expectedSum := 1.0 // seed
	for _, a := range amounts {
		expectedSum += a
	}
	var dbSum float64
	if err := config.DB.Model(&models.FinancialLedger{}).
		Where("unit_id = ? AND status = ?", unit.ID, "posted").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&dbSum).Error; err != nil {
		t.Fatalf("sum amounts: %v", err)
	}
	if dbSum != expectedSum {
		t.Fatalf("(C) amount mismatch: expected %v, got %v (successes=%d, total=%d)",
			expectedSum, dbSum, successes, total)
	}

	// Every goroutine that returned nil error must have a committed row.
	if successes != int(total)-1 {
		t.Fatalf("(C) successes=%d but DB has %d non-seed rows", successes, int(total)-1)
	}

	// Zero retries expected: the atomic counter guarantees correctness;
	// any error here indicates a real failure, not a transient conflict.
	if successes != N {
		t.Fatalf("expected %d successful appends with zero retries, got %d", N, successes)
	}
}
