package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	AccountabilityEventWeeklyDue       = "weekly_due"
	AccountabilityEventWeeklyOverdue   = "weekly_overdue"
	AccountabilityEventAdminEscalation = "admin_escalation"
)

type CaseAccountabilityEvent struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_case_accountability_event" json:"caseId"`
	OfficerID *uuid.UUID     `gorm:"type:uuid;index" json:"officerId,omitempty"`
	EventType string         `gorm:"not null;uniqueIndex:idx_case_accountability_event" json:"eventType"`
	WeekStart time.Time      `gorm:"not null;uniqueIndex:idx_case_accountability_event" json:"weekStart"`
	SentAt    time.Time      `gorm:"not null" json:"sentAt"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CaseAccountabilityEvent) TableName() string {
	return "case_accountability_events"
}
