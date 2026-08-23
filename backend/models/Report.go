package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Report struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReportedBy uuid.UUID      `gorm:"type:uuid;not null" json:"reportedBy"`
	TargetType string         `gorm:"not null" json:"targetType"` // case, sos, announcement, user
	TargetID   string         `gorm:"not null" json:"targetId"`
	Reason     string         `gorm:"type:text;not null" json:"reason"`
	Status     string         `gorm:"default:pending" json:"status"` // pending, reviewed, dismissed
	ReviewedBy *uuid.UUID     `gorm:"type:uuid" json:"reviewedBy,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
