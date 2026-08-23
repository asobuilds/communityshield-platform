package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Notification struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Title     string         `gorm:"not null" json:"title"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	Type      string         `json:"type"`
	Status    string         `gorm:"default:unread" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}
