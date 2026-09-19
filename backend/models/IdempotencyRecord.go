package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IdempotencyRecord stores the outcome of a request keyed by an
// Idempotency-Key header, so a retry replays the original response
// instead of re-executing the handler.
type IdempotencyRecord struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Key          string         `gorm:"type:varchar(128);not null;uniqueIndex:idx_idempotency_key_user" json:"key"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_idempotency_key_user" json:"userId"`
	Method       string         `gorm:"type:varchar(10);not null" json:"method"`
	Path         string         `gorm:"type:varchar(255);not null" json:"path"`
	Status       string         `gorm:"type:varchar(20);not null;default:in_progress" json:"status"`
	StatusCode   int            `gorm:"default:0" json:"statusCode"`
	ResponseBody string         `gorm:"type:text" json:"responseBody,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	ExpiresAt    time.Time      `gorm:"not null;index:idx_idempotency_expires" json:"expiresAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
