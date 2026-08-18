package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BankAccount represents a unit's bank account for receiving funds
type BankAccount struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID         uuid.UUID      `gorm:"type:uuid;not null;unique" json:"unitId"`
	BankName       string         `gorm:"not null" json:"bankName"`
	AccountNumber  string         `gorm:"not null" json:"accountNumber"`
	AccountName    string         `gorm:"not null" json:"accountName"`
	BankCode       string         `json:"bankCode"`
	Branch         string         `json:"branch"`
	SwiftCode      string         `json:"swiftCode"`
	IsVerified     bool           `gorm:"default:false" json:"isVerified"`
	IsDefault      bool           `gorm:"default:false" json:"isDefault"`
	Status         string         `gorm:"default:active" json:"status"` // active, inactive, suspended
	Notes          string         `json:"notes"`
	CreatedBy      uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Unit     SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Creator  User         `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (BankAccount) TableName() string {
	return "bank_accounts"
}

// Donation represents a donation made to a unit
type Donation struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID         uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	DonorName      string         `gorm:"not null" json:"donorName"`
	DonorEmail     string         `json:"donorEmail"`
	DonorPhone     string         `json:"donorPhone"`
	Amount         float64        `gorm:"not null" json:"amount"`
	Currency       string         `gorm:"default:NGN" json:"currency"`
	PaymentMethod  string         `json:"paymentMethod"` // bank_transfer, cash, mobile_money
	Reference      string         `json:"reference"`
	Status         string         `gorm:"default:pending" json:"status"` // pending, confirmed, failed
	DonationDate   time.Time      `json:"donationDate"`
	Notes          string         `json:"notes"`
	ReceiptURL     string         `json:"receiptUrl"`
	ConfirmedBy    *uuid.UUID     `gorm:"type:uuid" json:"confirmedBy,omitempty"`
	ConfirmedAt    *time.Time     `json:"confirmedAt,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Unit        SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Confirmer   User         `gorm:"foreignKey:ConfirmedBy" json:"confirmer,omitempty"`
}

func (Donation) TableName() string {
	return "donations"
}