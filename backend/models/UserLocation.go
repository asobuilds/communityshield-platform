package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserLocation struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	Latitude   float64        `gorm:"not null" json:"latitude"`
	Longitude  float64        `gorm:"not null" json:"longitude"`
	Accuracy   float64        `json:"accuracy"`
	RecordedAt time.Time      `gorm:"not null;index" json:"recordedAt"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UserLocation) TableName() string {
	return "user_locations"
}
