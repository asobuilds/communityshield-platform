package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	CaseID    uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	Title     string         `gorm:"not null" json:"title"`
	Message   string         `gorm:"not null" json:"message"`
	Read      bool           `gorm:"default:false" json:"read"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}