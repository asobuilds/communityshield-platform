package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SOSAlert struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      *uuid.UUID     `gorm:"type:uuid" json:"userId,omitempty"`
	UnitID      *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	Latitude    float64        `gorm:"not null" json:"latitude"`
	Longitude   float64        `gorm:"not null" json:"longitude"`
	Description string         `json:"description"`
	Status      string         `gorm:"default:pending" json:"status"`
	Anonymous   bool           `gorm:"default:false" json:"anonymous"`
	DeviceID    string         `gorm:"index" json:"deviceId"`
	IPAddress   string         `json:"ipAddress"`
	UserAgent   string         `json:"userAgent"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}