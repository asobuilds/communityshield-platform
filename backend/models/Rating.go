package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Rating struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	TargetID   uuid.UUID      `gorm:"type:uuid;not null" json:"targetId"`
	TargetType string         `gorm:"not null" json:"targetType"` // "unit" or "officer"
	Rating     int            `gorm:"not null" json:"rating"`
	Comment    string         `json:"comment"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Rating) TableName() string {
	return "ratings"
}