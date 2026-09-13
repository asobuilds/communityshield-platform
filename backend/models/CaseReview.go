package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	CaseReviewDecisionDeescalate     = "deescalate"
	CaseReviewDecisionRequestChanges = "request_changes"
	CaseReviewDecisionApprove        = "approve"
)

type CaseReview struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_case_review_admin" json:"caseId"`
	AdminID  uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_case_review_admin" json:"adminId"`
	Decision string    `gorm:"not null;index" json:"decision"`
	Comment  string    `gorm:"type:text;not null" json:"comment"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CaseReview) TableName() string {
	return "case_reviews"
}
