package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Rating struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID    uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	CaseID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"caseId"`
	Rating    int            `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment   string         `json:"comment"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}