package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BankAccount represents a unit's bank account for receiving funds
type BankAccount struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID        uuid.UUID      `gorm:"type:uuid;not null;unique" json:"unitId"`
	BankName      string         `gorm:"not null" json:"bankName"`
	AccountNumber string         `gorm:"not null" json:"accountNumber"`
	AccountName   string         `gorm:"not null" json:"accountName"`
	BankCode      string         `json:"bankCode"`
	Branch        string         `json:"branch"`
	SwiftCode     string         `json:"swiftCode"`
	IsVerified    bool           `gorm:"default:false" json:"isVerified"`
	IsDefault     bool           `gorm:"default:false" json:"isDefault"`
	Status        string         `gorm:"default:active" json:"status"`
	Notes         string         `json:"notes"`
	CreatedBy     uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Unit    SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Creator User         `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (BankAccount) TableName() string {
	return "bank_accounts"
}