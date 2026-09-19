package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"

	"security-solution/config"
	"security-solution/models"
)

type SchedulerService struct {
	Election   *ElectionService
	Revocation *RevocationService
}

func NewSchedulerService() *SchedulerService {
	return &SchedulerService{
		Election:   NewElectionService(),
		Revocation: NewRevocationService(),
	}
}

// RunOnce acquires a Postgres advisory lock so only one backend instance
// runs scheduled jobs at a time, then performs one pass of all jobs.
// If the lock is held elsewhere, this call is a no-op.
func (s *SchedulerService) RunOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sqlDB, err := config.DB.DB()
	if err != nil {
		log.Printf("scheduler: cannot get sql.DB: %v", err)
		return
	}

	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		log.Printf("scheduler: cannot acquire connection: %v", err)
		return
	}
	defer conn.Close()

	// Advisory lock ID: any fixed int64. Pick one and never change it.
	const schedulerLockID int64 = 8420194710293847

	var got bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", schedulerLockID).Scan(&got); err != nil {
		log.Printf("scheduler: advisory lock query failed: %v", err)
		return
	}
	if !got {
		// Another instance is running the scheduler right now.
		return
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", schedulerLockID)
	}()

	s.CloseExpiredElections()
	s.ExpireStaleInvites()
	s.CleanupExpiredSessions()
	s.CloseFinishedFinancialYears()
	s.RecomputeRankings()
	s.AnnounceOfficerOfTheWeek()
	s.EscalateStaleAppeals()
	s.RecomputeMinorStatuses()
}

// CloseExpiredElections finalizes any open UnitAdminElection whose voting window has passed.
func (s *SchedulerService) CloseExpiredElections() {
	now := time.Now().UTC()

	var elections []models.UnitAdminElection
	if err := config.DB.
		Where("status = ? AND voting_ends_at IS NOT NULL AND voting_ends_at <= ?", "open", now).
		Find(&elections).Error; err != nil {
		log.Printf("scheduler: failed to load expired elections: %v", err)
		return
	}

	for _, e := range elections {
		var err error
		if e.ElectionType == "head_admin" {
			err = s.Election.FinalizeHeadAdminElection(e.ID)
		} else {
			err = s.Election.CloseAdminElection(e.ID)
		}
		if err != nil {
			log.Printf("scheduler: failed to close election %s: %v", e.ID, err)
		}
	}
}

// ExpireStaleInvites marks pending invites as expired once their expiry date passes.
func (s *SchedulerService) ExpireStaleInvites() {
	now := time.Now().UTC()

	if err := config.DB.
		Model(&models.UnitInvite{}).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at <= ?", "pending", now).
		Update("status", "expired").Error; err != nil {
		log.Printf("scheduler: failed to expire invites: %v", err)
	}
}

// Start runs the scheduler loop every interval until stopped. Call from main.
func (s *SchedulerService) Start(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			s.RunOnce()
		}
	}()
}

func (s *SchedulerService) CleanupExpiredSessions() {
	_ = NewSessionService().CleanupExpired()
	_ = NewRefreshTokenService().CleanupExpired()
	_ = NewTokenService().CleanupExpired()
	_ = config.DB.Where("expires_at < ?", time.Now().UTC()).Delete(&models.IdempotencyRecord{}).Error
}

// CloseFinishedFinancialYears closes the previous calendar year for every
// unit that has ledger activity but no UnitFinancialYear record yet.
// Runs daily; idempotent — running twice in one day is harmless.
func (s *SchedulerService) CloseFinishedFinancialYears() {
	currentYear := time.Now().UTC().Year()
	prevYear := currentYear - 1

	type unitRow struct {
		UnitID uuid.UUID
	}
	var units []unitRow
	config.DB.Model(&models.FinancialLedger{}).
		Select("DISTINCT unit_id").
		Where("year = ?", prevYear).
		Scan(&units)

	ledgerSvc := NewLedgerService()
	for _, u := range units {
		var existing models.UnitFinancialYear
		if err := config.DB.Where("unit_id = ? AND year = ?", u.UnitID, prevYear).
			First(&existing).Error; err == nil {
			continue
		}
		_, _ = ledgerSvc.CloseYear(u.UnitID, prevYear, uuid.Nil)
	}
}

// RecomputeRankings refreshes officer and unit ranking scores from current ratings.
func (s *SchedulerService) RecomputeRankings() {
	_ = NewRankingService().RecomputeAll()
}

// EscalateStaleAppeals flags unresolved appeals past the 30-day threshold and
// notifies all super admins. Delegates to the appeal service.
func (s *SchedulerService) EscalateStaleAppeals() {
	NewAppealService().EscalateStaleAppeals()
}

// AnnounceOfficerOfTheWeek picks the top-rated officer of the past 7 days
// and creates a news post. Runs at most once per week — guarded by checking
// whether an OOW post already exists in the last 7 days.
// Only fires on Mondays after 09:00 UTC.
func (s *SchedulerService) AnnounceOfficerOfTheWeek() {
	now := time.Now().UTC()
	if now.Weekday() != time.Monday || now.Hour() < 9 {
		return
	}

	// Idempotency guard — has an OOW post gone out in the last 7 days?
	var recentCount int64
	config.DB.Model(&models.News{}).
		Where("category = ? AND created_at > ?", "officer_of_the_week", now.AddDate(0, 0, -7)).
		Count(&recentCount)
	if recentCount > 0 {
		return
	}

	// Count ratings for each officer in the last 7 days; require >= 3 new.
	type row struct {
		TargetID uuid.UUID
		Sum      float64
		Count    int64
	}
	var rows []row
	config.DB.Model(&models.Rating{}).
		Select("target_id, COALESCE(SUM(rating), 0) as sum, COUNT(*) as count").
		Where("target_type = ? AND status = ? AND created_at > ?",
			"officer", "active", now.AddDate(0, 0, -7)).
		Group("target_id").
		Having("COUNT(*) >= 3").
		Scan(&rows)

	if len(rows) == 0 {
		return
	}

	// Pick the highest weekly Bayesian score
	var winner row
	var bestScore float64 = -1
	for _, r := range rows {
		bayes := BayesianScore(r.Sum, r.Count)
		if bayes > bestScore {
			bestScore = bayes
			winner = r
		}
	}

	var officer models.Officer
	if err := config.DB.First(&officer, "id = ?", winner.TargetID).Error; err != nil {
		return
	}

	var unit models.SecurityUnit
	config.DB.First(&unit, "id = ?", officer.UnitID)

	title := "Officer of the Week: " + officer.Name
	content := fmt.Sprintf(
		"Congratulations to %s of %s.\n\nWeekly rating: %.2f★ from %d verified ratings this week.\n\nThank you for your service.",
		officer.Name, unit.Name, bestScore, winner.Count,
	)

	news := models.News{
		Title:       title,
		Content:     content,
		Source:      "WardGuard",
		Author:      "WardGuard",
		Category:    "officer_of_the_week",
		Location:    unit.Name,
		Sentiment:   "positive",
		ThreatLevel: "low",
		Status:      "published",
		PublishedAt: now,
		CreatedBy:   uuid.Nil,
	}
	config.DB.Create(&news)

	// Notify the officer
	notification := models.Notification{
		UserID:  officer.ID,
		Title:   "You are Officer of the Week!",
		Message: fmt.Sprintf("Recognized platform-wide for %d verified ratings this week (%.2f★).", winner.Count, bestScore),
		Type:    "officer_of_the_week",
		Status:  "unread",
	}
	config.DB.Create(&notification)
}

// RecomputeMinorStatuses rolls over minors who have aged out (and adults who
// somehow drifted into "minor") based on the stored DateOfBirth. Skipped for
// users with a super-admin-granted MinorExceptionGranted to avoid surprises.
func (s *SchedulerService) RecomputeMinorStatuses() {
	NewAgeService().RecomputeMinorStatuses()
}