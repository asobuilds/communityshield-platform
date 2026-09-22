package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlatformDonation represents a donation made to Nativity Guard itself
// (not to a unit). Confirmed by a super admin. Optionally public.
type PlatformDonation struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DonorID     *uuid.UUID     `gorm:"type:uuid;index:idx_platform_donation_donor" json:"donorId,omitempty"`
	DonorName   string         `gorm:"type:varchar(160);not null" json:"donorName"`
	DonorEmail  string         `gorm:"type:varchar(160)" json:"donorEmail,omitempty"`
	DonorPhone  string         `gorm:"type:varchar(40)" json:"donorPhone,omitempty"`
	Amount      float64        `gorm:"not null" json:"amount"`
	Currency    string         `gorm:"type:varchar(8);not null;default:'NGN'" json:"currency"`
	Method      string         `gorm:"type:varchar(32);not null;default:'bank_transfer'" json:"method"`
	ReferenceID string         `gorm:"type:varchar(120)" json:"referenceId,omitempty"`
	Message     string         `gorm:"type:text" json:"message,omitempty"`
	IsPublic    bool           `gorm:"default:true" json:"isPublic"`
	Status      string         `gorm:"type:varchar(16);not null;default:'pending';index:idx_platform_donation_status" json:"status"`
	ConfirmedBy *uuid.UUID     `gorm:"type:uuid" json:"confirmedBy,omitempty"`
	ConfirmedAt *time.Time     `json:"confirmedAt,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PlatformDonation) TableName() string {
	return "platform_donations"
}
