package services

import (
	"context"
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