package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

const (
    UnitVerificationPending      = "pending"
    UnitVerificationReadyForReview = "ready_for_review"
    UnitVerificationUnderReview = "under_review"
    UnitVerificationVerified     = "verified"
    UnitVerificationRejected     = "rejected"
)

type UnitVerification struct {
    ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

    UnitID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"unitId"`

    Status string `gorm:"not null;default:pending;index" json:"status"`

    AdminCount int `gorm:"not null;default:0" json:"adminCount"`

    MinimumAdmins int `gorm:"not null;default:5" json:"minimumAdmins"`

    SubmittedAt *time.Time `json:"submittedAt,omitempty"`

    ReviewedAt *time.Time `json:"reviewedAt,omitempty"`

    ReviewedBy *uuid.UUID `gorm:"type:uuid" json:"reviewedBy,omitempty"`

    RejectionReason string `json:"rejectionReason,omitempty"`

    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    Unit     UserUnit `gorm:"foreignKey:UnitID" json:"-"`
    Reviewer *User    `gorm:"foreignKey:ReviewedBy" json:"-"`
}

func (UnitVerification) TableName() string {
    return "unit_verifications"
}
