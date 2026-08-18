package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SystemSettings struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Key         string         `gorm:"unique;not null" json:"key"`
	Value       string         `gorm:"type:text" json:"value"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
	IsPublic    bool           `gorm:"default:false" json:"isPublic"`
	UpdatedBy   *uuid.UUID     `gorm:"type:uuid" json:"updatedBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Updater User `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

func (SystemSettings) TableName() string {
	return "system_settings"
}