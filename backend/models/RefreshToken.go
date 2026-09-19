package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;index:idx_refresh_token_user" json:"userId"`
	TokenHash  string         `gorm:"type:varchar(64);not null;uniqueIndex:idx_refresh_token_hash" json:"-"`
	SessionID  *uuid.UUID     `gorm:"type:uuid;index:idx_refresh_token_session" json:"sessionId,omitempty"`
	ExpiresAt  time.Time      `gorm:"not null;index:idx_refresh_token_expires" json:"expiresAt"`
	RevokedAt  *time.Time     `json:"revokedAt,omitempty"`
	ReplacedBy *uuid.UUID     `gorm:"type:uuid" json:"replacedBy,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
