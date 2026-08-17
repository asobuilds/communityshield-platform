package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Announcement struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID      *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	Title       string         `gorm:"not null" json:"title"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Type        string         `gorm:"not null;default:news" json:"type"` // news, alert, warning
	Severity    string         `gorm:"not null;default:medium" json:"severity"` // low, medium, high, critical
	Latitude    *float64       `json:"latitude,omitempty"`
	Longitude   *float64       `json:"longitude,omitempty"`
	RadiusKM    int            `gorm:"default:0" json:"radiusKm"` // 0 = no filter
	IsPublic    bool           `gorm:"default:true" json:"isPublic"`
	ExpiresAt   *time.Time     `json:"expiresAt,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}