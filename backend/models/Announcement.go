package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Announcement struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title     string         `gorm:"not null" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	UnitID    *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	CreatedBy uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	Status    string         `gorm:"default:active" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Author User `gorm:"foreignKey:CreatedBy" json:"author,omitempty"`
}

func (Announcement) TableName() string {
	return "announcements"
}
