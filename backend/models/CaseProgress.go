package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type CaseProgress struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID      uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	OfficerID   uuid.UUID      `gorm:"type:uuid;not null" json:"officerId"`
	Action      string         `gorm:"not null" json:"action"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
