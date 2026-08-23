package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GovernmentIDVerification struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`

	IDType   string `gorm:"not null" json:"idType"`
	IDNumber string `gorm:"not null" json:"idNumber"`

	DocumentURL string `json:"documentUrl"`

	Status string `gorm:"not null;default:pending;index" json:"status"` // pending, approved, rejected

	SubmittedAt time.Time  `json:"submittedAt"`
	ReviewedAt  *time.Time `json:"reviewedAt,omitempty"`
	ReviewedBy  *uuid.UUID `gorm:"type:uuid" json:"reviewedBy,omitempty"`

	RejectionReason string `json:"rejectionReason,omitempty"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (GovernmentIDVerification) TableName() string {
	return "government_id_verifications"
}
