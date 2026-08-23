package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type PushSubscription struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	Endpoint  string         `gorm:"unique;not null" json:"endpoint"`
	Keys      string         `gorm:"type:jsonb;not null" json:"keys"` // JSON object {p256dh, auth}
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
