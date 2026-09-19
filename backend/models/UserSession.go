package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSession struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_session_user" json:"userId"`
	JTI          string         `gorm:"type:varchar(64);not null;uniqueIndex:idx_user_session_jti" json:"jti"`
	Device       string         `gorm:"type:varchar(120)" json:"device,omitempty"`
	UserAgent    string         `gorm:"type:text" json:"userAgent,omitempty"`
	IPAddress    string         `gorm:"type:varchar(45)" json:"ipAddress,omitempty"`
	LastSeenAt   time.Time      `gorm:"not null;index:idx_user_session_last_seen" json:"lastSeenAt"`
	ExpiresAt    time.Time      `gorm:"not null;index:idx_user_session_expires" json:"expiresAt"`
	RevokedAt    *time.Time     `json:"revokedAt,omitempty"`
	RevokedReason *string       `gorm:"type:text" json:"revokedReason,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
