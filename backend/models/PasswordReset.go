package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type PasswordReset struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Token     string         `gorm:"unique;not null" json:"token"`
	ExpiresAt time.Time      `json:"expiresAt"`
	Used      bool           `gorm:"default:false" json:"used"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (PasswordReset) TableName() string {
	return "password_resets"
}
