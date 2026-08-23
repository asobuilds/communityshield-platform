package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type CaseTemplate struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID      *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Fields      string         `gorm:"type:text" json:"fields"` // JSON schema
	Priority    string         `gorm:"default:medium" json:"priority"`
	Status      string         `gorm:"default:active" json:"status"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CaseTemplate) TableName() string {
	return "case_templates"
}
