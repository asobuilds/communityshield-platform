package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type PushToken struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Token     string         `gorm:"not null" json:"token"`
	Device    string         `json:"device"`
	Active    bool           `gorm:"default:true" json:"active"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (PushToken) TableName() string {
	return "push_tokens"
}
