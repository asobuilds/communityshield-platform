package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseWeeklyUpdate struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_case_week_officer" json:"caseId"`
	OfficerID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_case_week_officer" json:"officerId"`

	WeekStart time.Time `gorm:"not null;uniqueIndex:idx_case_week_officer" json:"weekStart"`
	WeekEnd   time.Time `gorm:"not null" json:"weekEnd"`

	Summary            string `gorm:"type:text;not null" json:"summary"`
	Investigation      string `gorm:"type:text;not null" json:"investigation"`
	ActionsTaken       string `gorm:"type:text" json:"actionsTaken"`
	Findings           string `gorm:"type:text" json:"findings"`
	EvidenceSummary    string `gorm:"type:text" json:"evidenceSummary"`
	OutstandingActions string `gorm:"type:text" json:"outstandingActions"`
	NextSteps          string `gorm:"type:text" json:"nextSteps"`

	// Only safe, non-sensitive information is exposed to the reporter.
	CitizenVisible bool `gorm:"default:true" json:"citizenVisible"`

	SubmittedAt *time.Time `json:"submittedAt,omitempty"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CaseWeeklyUpdate) TableName() string {
	return "case_weekly_updates"
}
