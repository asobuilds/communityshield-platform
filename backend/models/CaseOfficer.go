package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type CaseOfficer struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID    uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	OfficerID uuid.UUID      `gorm:"type:uuid;not null" json:"officerId"`
	Role      string         `gorm:"default:investigator" json:"role"` // primary, investigator, support
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
