package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseTimeline struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID      uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Action      string         `gorm:"not null" json:"action"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `json:"status"`
	Metadata    string         `gorm:"type:text" json:"metadata"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Case Case `gorm:"foreignKey:CaseID" json:"case,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (CaseTimeline) TableName() string {
	return "case_timelines"
}