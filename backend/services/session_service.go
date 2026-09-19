package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"security-solution/config"
	"security-solution/models"
)

type SessionService struct{}

func NewSessionService() *SessionService {
	return &SessionService{}
}

func (s *SessionService) Create(userID uuid.UUID, jti string, device string, userAgent string, ipAddress string, expiresAt time.Time) (*models.UserSession, error) {
	if jti == "" {
		return nil, errors.New("jti is required")
	}

	session := models.UserSession{
		UserID:     userID,
		JTI:        jti,
		Device:     device,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
		LastSeenAt: time.Now().UTC(),
		ExpiresAt:  expiresAt,
	}

	if err := config.DB.Create(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *SessionService) Touch(jti string) {
	if jti == "" {
		return
	}
	config.DB.Model(&models.UserSession{}).
		Where("jti = ? AND revoked_at IS NULL AND last_seen_at < ?", jti, time.Now().UTC().Add(-1*time.Minute)).
		Update("last_seen_at", time.Now().UTC())
}

func (s *SessionService) ListActive(userID uuid.UUID) ([]models.UserSession, error) {
	var sessions []models.UserSession
	err := config.DB.
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now().UTC()).
		Order("last_seen_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (s *SessionService) RevokeOne(userID uuid.UUID, jti string, reason string) error {
	var session models.UserSession
	if err := config.DB.Where("user_id = ? AND jti = ?", userID, jti).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("session not found")
		}
		return err
	}

	now := time.Now().UTC()
	session.RevokedAt = &now
	session.RevokedReason = &reason
	return config.DB.Save(&session).Error
}

func (s *SessionService) RevokeAll(userID uuid.UUID, reason string) error {
	now := time.Now().UTC()
	return config.DB.Model(&models.UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]interface{}{
			"revoked_at":     now,
			"revoked_reason": reason,
		}).Error
}

func (s *SessionService) CleanupExpired() error {
	return config.DB.
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&models.UserSession{}).Error
}
