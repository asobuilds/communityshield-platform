package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

type RefreshTokenService struct{}

func NewRefreshTokenService() *RefreshTokenService {
	return &RefreshTokenService{}
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *RefreshTokenService) Issue(userID uuid.UUID, sessionID *uuid.UUID) (string, *models.RefreshToken, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	tokenRaw := hex.EncodeToString(raw)

	record := models.RefreshToken{
		UserID:    userID,
		SessionID: sessionID,
		TokenHash: hashToken(tokenRaw),
		ExpiresAt: time.Now().UTC().AddDate(0, 0, 30),
	}
	if err := config.DB.Create(&record).Error; err != nil {
		return "", nil, err
	}
	return tokenRaw, &record, nil
}

func (s *RefreshTokenService) Rotate(oldRaw string) (string, *models.RefreshToken, error) {
	var existing models.RefreshToken
	if err := config.DB.Where("token_hash = ?", hashToken(oldRaw)).First(&existing).Error; err != nil {
		return "", nil, errors.New("refresh token not found")
	}
	if existing.RevokedAt != nil {
		return "", nil, errors.New("refresh token has been revoked")
	}
	if time.Now().UTC().After(existing.ExpiresAt) {
		return "", nil, errors.New("refresh token has expired")
	}

	newRaw, newRecord, err := s.Issue(existing.UserID, existing.SessionID)
	if err != nil {
		return "", nil, err
	}

	now := time.Now().UTC()
	existing.RevokedAt = &now
	existing.ReplacedBy = &newRecord.ID
	if err := config.DB.Save(&existing).Error; err != nil {
		return "", nil, err
	}

	return newRaw, newRecord, nil
}

func (s *RefreshTokenService) RevokeAllForUser(userID uuid.UUID) error {
	now := time.Now().UTC()
	return config.DB.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (s *RefreshTokenService) CleanupExpired() error {
	return config.DB.
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&models.RefreshToken{}).Error
}

func (s *RefreshTokenService) RevokeBySessionID(sessionID uuid.UUID, reason string) error {
	now := time.Now().UTC()
	return config.DB.Model(&models.RefreshToken{}).
		Where("session_id = ? AND revoked_at IS NULL", sessionID).
		Update("revoked_at", now).Error
}
