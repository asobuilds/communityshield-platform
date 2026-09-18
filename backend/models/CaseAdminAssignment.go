package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseAdminAssignment struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID           uuid.UUID      `gorm:"type:uuid;not null;index:idx_case_admin_assignment_case;uniqueIndex:idx_case_admin_assignment_case_admin" json:"caseId"`
	AdminID          uuid.UUID      `gorm:"type:uuid;not null;index:idx_case_admin_assignment_admin;uniqueIndex:idx_case_admin_assignment_case_admin" json:"adminId"`
	SubmittedBy      uuid.UUID      `gorm:"type:uuid;not null;index:idx_case_admin_assignment_submitted_by" json:"submittedBy"`
	Status           string         `gorm:"not null;default:pending;index:idx_case_admin_assignment_status" json:"status"`
	SubmittedAt      time.Time      `gorm:"not null;default:now()" json:"submittedAt"`
	ApprovedAt       *time.Time     `json:"approvedAt,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
