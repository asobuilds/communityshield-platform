package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID          uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Amount          float64        `gorm:"not null" json:"amount"`
	Type            string         `gorm:"not null" json:"type"`
	Description     string         `gorm:"not null" json:"description"`
	TransactionDate time.Time      `gorm:"not null" json:"transactionDate"`
	InitiatedBy     uuid.UUID      `gorm:"type:uuid;not null" json:"initiatedBy"`
	ApprovedBy      *uuid.UUID     `gorm:"type:uuid" json:"approvedBy,omitempty"`
	Status          string         `gorm:"default:pending" json:"status"`
	ReferenceID     string         `json:"referenceId,omitempty"`
	PaymentMethod   string         `json:"paymentMethod"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}