package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// Donation represents a donation made to a unit
type Donation struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID        uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	DonorName     string         `gorm:"not null" json:"donorName"`
	DonorEmail    string         `json:"donorEmail"`
	DonorPhone    string         `json:"donorPhone"`
	Amount        float64        `gorm:"not null" json:"amount"`
	Currency      string         `gorm:"default:NGN" json:"currency"`
	PaymentMethod string         `json:"paymentMethod"`
	Reference     string         `json:"reference"`
	Status        string         `gorm:"default:pending" json:"status"`
	DonationDate  time.Time      `json:"donationDate"`
	Notes         string         `json:"notes"`
	ReceiptURL    string         `json:"receiptUrl"`
	ConfirmedBy   *uuid.UUID     `gorm:"type:uuid" json:"confirmedBy,omitempty"`
	ConfirmedAt   *time.Time     `json:"confirmedAt,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Unit      SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Confirmer User         `gorm:"foreignKey:ConfirmedBy" json:"confirmer,omitempty"`
}

func (Donation) TableName() string {
	return "donations"
}
