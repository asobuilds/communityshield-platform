package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RevokedToken stores a JWT's unique id (jti) once it has been revoked.
// The middleware checks this table on every authenticated request.
type RevokedToken struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	JTI       string         `gorm:"type:varchar(64);not null;uniqueIndex:idx_revoked_token_jti" json:"jti"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_revoked_token_user" json:"userId"`
	ExpiresAt time.Time      `gorm:"not null;index:idx_revoked_token_expires" json:"expiresAt"`
	Reason    string         `gorm:"type:varchar(64);not null;default:'logout'" json:"reason"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}