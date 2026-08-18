package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OTP struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Code      string         `gorm:"not null" json:"code"`
	Type      string         `json:"type"`
	Attempts  int            `gorm:"default:0" json:"attempts"`
	ExpiresAt time.Time      `json:"expiresAt"`
	Verified  bool           `gorm:"default:false" json:"verified"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OTP) TableName() string {
	return "otps"
}