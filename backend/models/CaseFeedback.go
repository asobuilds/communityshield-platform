package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type CaseFeedback struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID     uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Rating     int            `gorm:"not null" json:"rating"` // 1-5
	Comment    string         `gorm:"type:text" json:"comment"`
	Categories string         `gorm:"type:text" json:"categories"` // JSON array of categories
	IsPublic   bool           `gorm:"default:false" json:"isPublic"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Case Case `gorm:"foreignKey:CaseID" json:"case,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (CaseFeedback) TableName() string {
	return "case_feedback"
}
