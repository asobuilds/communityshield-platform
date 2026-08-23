package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Alert struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title     string         `gorm:"not null" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Type      string         `gorm:"not null" json:"type"` // security, weather, community, warning
	Location  string         `json:"location"`
	Severity  string         `gorm:"default:medium" json:"severity"` // low, medium, high, critical
	Status    string         `gorm:"default:active" json:"status"`   // active, expired, resolved
	CreatedBy uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	ExpiresAt *time.Time     `json:"expiresAt"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Author User `gorm:"foreignKey:CreatedBy" json:"author,omitempty"`
}

func (Alert) TableName() string {
	return "alerts"
}
