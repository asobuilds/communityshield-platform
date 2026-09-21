//go:build integration

package handlers

import (
	"testing"
	"time"

	"security-solution/config"
	"security-solution/internal/testutil"
	"security-solution/services"
)

// TestRatings_BayesianPriorDominatesAtLowCount — with 0 ratings, score ≈ prior.
func TestRatings_BayesianPriorDominatesAtLowCount(t *testing.T) {
	testutil.TruncateAll(t)

	// 1 five-star rating → score should be near 3.1, not 5.0
	score := services.BayesianScore(5, 1)
	if score >= 4.0 {
		t.Fatalf("expected prior to dominate at 1 rating, got %.2f", score)
	}
	if score <= 3.0 {
		t.Fatalf("expected score to rise above prior with a 5-star, got %.2f", score)
	}
}

// TestRatings_BayesianConvergesAtHighCount — with many 5-star ratings, score → 5.0.
func TestRatings_BayesianConvergesAtHighCount(t *testing.T) {
	testutil.TruncateAll(t)

	score := services.BayesianScore(500*5, 500)
	if score < 4.8 {
		t.Fatalf("expected convergence near 5.0 at 500 ratings, got %.2f", score)
	}
}

// TestRatings_TierBoundaries — tier function maps correctly.
func TestRatings_TierBoundaries(t *testing.T) {
	testutil.TruncateAll(t)

	cases := []struct {
		score float64
		count int64
		tier  string
	}{
		{4.9, 1, "unranked"},   // score too good, count too low
		{3.5, 4, "unranked"},   // count below bronze threshold
		{3.5, 5, "bronze"},
		{4.0, 20, "silver"},
		{4.3, 50, "gold"},
		{4.6, 150, "platinum"},
		{4.8, 500, "diamond"},
	}
	for _, c := range cases {
		got := services.TierForOfficer(c.score, c.count)
		if got != c.tier {
			t.Fatalf("score=%.1f count=%d: expected %s, got %s", c.score, c.count, c.tier, got)
		}
	}
}

// TestRatings_OnlyReporterMayRate — a non-reporter attempt is rejected.
func TestRatings_OnlyReporterMayRate(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	reporter := testutil.MakeUser(t, "citizen")
	officer := testutil.MakeUser(t, "officer")
	testutil.MakeMembership(t, officer, unit, "officer", false)

	c := testutil.MakeCase(t, reporter, unit)
	now := time.Now().UTC()
	config.DB.Model(c).Updates(map[string]interface{}{
		"assigned_to":          officer.ID,
		"status":               "closed",
		"closed_at":            now,
		"closed_via_platform":  true,
	})

	svc := services.NewRatingService()

	// A random non-reporter tries to rate
	stranger := testutil.MakeUser(t, "citizen")
	_, err := svc.SubmitRating(services.SubmitRatingRequest{
		RaterUserID: stranger.ID,
		CaseID:      c.ID,
		TargetID:    officer.ID,
		TargetType:  "officer",
		Rating:      5,
		Comment:     "nice",
	})
	if err == nil {
		t.Fatalf("expected non-reporter to be rejected")
	}
}

// TestRatings_OnlyClosedCasesMayBeRated — open case rejection.
func TestRatings_OnlyClosedCasesMayBeRated(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	reporter := testutil.MakeUser(t, "citizen")
	officer := testutil.MakeUser(t, "officer")
	testutil.MakeMembership(t, officer, unit, "officer", false)

	c := testutil.MakeCase(t, reporter, unit) // status pending
	config.DB.Model(c).Update("assigned_to", officer.ID)

	svc := services.NewRatingService()
	_, err := svc.SubmitRating(services.SubmitRatingRequest{
		RaterUserID: reporter.ID,
		CaseID:      c.ID,
		TargetID:    officer.ID,
		TargetType:  "officer",
		Rating:      5,
	})
	if err == nil {
		t.Fatalf("expected open case to be rejected")
	}
}

// TestRatings_OnlyPlatformClosedCasesMayBeRated — off-platform closure rejection.
func TestRatings_OnlyPlatformClosedCasesMayBeRated(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	reporter := testutil.MakeUser(t, "citizen")
	officer := testutil.MakeUser(t, "officer")
	testutil.MakeMembership(t, officer, unit, "officer", false)

	c := testutil.MakeCase(t, reporter, unit)
	now := time.Now().UTC()
	// closed WITHOUT closed_via_platform
	config.DB.Model(c).Updates(map[string]interface{}{
		"assigned_to": officer.ID,
		"status":      "closed",
		"closed_at":   now,
	})

	svc := services.NewRatingService()
	_, err := svc.SubmitRating(services.SubmitRatingRequest{
		RaterUserID: reporter.ID,
		CaseID:      c.ID,
		TargetID:    officer.ID,
		TargetType:  "officer",
		Rating:      5,
	})
	if err == nil {
		t.Fatalf("expected off-platform closure to be rejected")
	}
}

// TestRatings_DuplicateRejected — one rating per (user, case, target).
func TestRatings_DuplicateRejected(t *testing.T) {
	testutil.TruncateAll(t)

	head := testutil.MakeUser(t, "unit_admin")
	unit := testutil.MakeUnit(t, head)
	reporter := testutil.MakeUser(t, "citizen")
	officer := testutil.MakeUser(t, "officer")
	testutil.MakeMembership(t, officer, unit, "officer", false)

	c := testutil.MakeCase(t, reporter, unit)
	now := time.Now().UTC()
	config.DB.Model(c).Updates(map[string]interface{}{
		"assigned_to":         officer.ID,
		"status":              "closed",
		"closed_at":           now,
		"closed_via_platform": true,
	})

	svc := services.NewRatingService()
	req := services.SubmitRatingRequest{
		RaterUserID: reporter.ID,
		CaseID:      c.ID,
		TargetID:    officer.ID,
		TargetType:  "officer",
		Rating:      5,
	}
	if _, err := svc.SubmitRating(req); err != nil {
		t.Fatalf("first rating: %v", err)
	}
	if _, err := svc.SubmitRating(req); err == nil {
		t.Fatalf("expected duplicate rating to be rejected")
	}
}