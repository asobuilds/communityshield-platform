package services

import (
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

// RunOnce performs one pass of all scheduled jobs. Safe to call repeatedly.
func (s *SchedulerService) RunOnce() {
	s.CloseExpiredElections()
	s.ExpireStaleInvites()
	s.CleanupExpiredSessions()
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
}