package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"security-solution/config"
	"security-solution/models"
)

type TokenService struct{}

func NewTokenService() *TokenService {
	return &TokenService{}
}

// Revoke marks a JWT (by its jti) as revoked until its natural expiry.
// Safe to call repeatedly — second call is a no-op.
func (s *TokenService) Revoke(jti string, userID uuid.UUID, expiresAt time.Time, reason string) error {
	if jti == "" {
		return errors.New("jti is required")
	}

	var existing models.RevokedToken
	err := config.DB.Where("jti = ?", jti).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if reason == "" {
		reason = "logout"
	}

	token := models.RevokedToken{
		JTI:       jti,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Reason:    reason,
	}
	return config.DB.Create(&token).Error
}

// IsRevoked returns true if the given jti is on the revocation list.
func (s *TokenService) IsRevoked(jti string) bool {
	if jti == "" {
		return false
	}
	var count int64
	config.DB.Model(&models.RevokedToken{}).
		Where("jti = ?", jti).
		Count(&count)
	return count > 0
}

// RevokeAllForUser revokes every currently-active token by tagging the user.
// Not a per-jti revoke — this is called on password change or account suspension,
// and the middleware handles the "revoked before" logic.
func (s *TokenService) RevokeAllForUser(userID uuid.UUID, reason string) error {
	// Store a sentinel row keyed by user id + timestamp-of-revocation.
	// The middleware checks: if token.iat < row.created_at → revoked.
	sentinel := models.RevokedToken{
		JTI:       "ALL:" + userID.String() + ":" + time.Now().UTC().Format("20060102150405"),
		UserID:    userID,
		ExpiresAt: time.Now().UTC().AddDate(0, 0, 30),
		Reason:    reason,
	}
	return config.DB.Create(&sentinel).Error
}

// IsUserRevokedAfter returns true if any sentinel row for this user was created
// after the given timestamp. Used for "revoke all tokens issued before X".
func (s *TokenService) IsUserRevokedAfter(userID uuid.UUID, issuedAt time.Time) bool {
	var count int64
	config.DB.Model(&models.RevokedToken{}).
		Where("user_id = ? AND jti LIKE ? AND created_at > ?", userID, "ALL:%", issuedAt).
		Count(&count)
	return count > 0
}

// CleanupExpired deletes revoked-token rows whose expires_at has passed.
// Called by the scheduler.
func (s *TokenService) CleanupExpired() error {
	return config.DB.
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&models.RevokedToken{}).Error
}