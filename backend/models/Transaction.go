package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Transaction struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID            uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Type              string         `gorm:"not null" json:"type"` // donation, bail, tax, gift, expense, salary, maintenance
	Category          string         `json:"category"`             // operational, capital, emergency, recurring
	Amount            float64        `gorm:"not null" json:"amount"`
	Description       string         `gorm:"not null" json:"description"`
	InitiatedBy       uuid.UUID      `gorm:"type:uuid;not null" json:"initiatedBy"`
	ApprovedBy        *uuid.UUID     `gorm:"type:uuid" json:"approvedBy,omitempty"`
	ApprovalCount     int            `gorm:"default:0" json:"approvalCount"`
	RequiredApprovals int            `gorm:"default:2" json:"requiredApprovals"`
	Status            string         `gorm:"default:pending" json:"status"` // pending, approved, rejected, completed
	TransactionDate   time.Time      `json:"transactionDate"`
	PaymentMethod     string         `json:"paymentMethod"` // cash, bank_transfer, mobile_money, cheque
	ReferenceID       string         `json:"referenceId,omitempty"`
	ReceiptURL        string         `json:"receiptUrl,omitempty"`
	Notes             string         `json:"notes"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	Unit      SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Initiator User         `gorm:"foreignKey:InitiatedBy" json:"initiator,omitempty"`
	Approver  User         `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

func (Transaction) TableName() string {
	return "transactions"
}

// TransactionApproval tracks multi-signature approvals for transactions
type TransactionApproval struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TransactionID uuid.UUID      `gorm:"type:uuid;not null" json:"transactionId"`
	ApproverID    uuid.UUID      `gorm:"type:uuid;not null" json:"approverId"`
	UnitID        uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Status        string         `gorm:"default:pending" json:"status"` // pending, approved, rejected
	Comment       string         `json:"comment"`
	ApprovedAt    *time.Time     `json:"approvedAt,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Transaction Transaction  `gorm:"foreignKey:TransactionID" json:"transaction,omitempty"`
	Approver    User         `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	Unit        SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
}

func (TransactionApproval) TableName() string {
	return "transaction_approvals"
}

// Budget represents unit budget allocations
type Budget struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID    uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Category  string         `gorm:"not null" json:"category"`
	Amount    float64        `gorm:"not null" json:"amount"`
	Spent     float64        `gorm:"default:0" json:"spent"`
	Remaining float64        `gorm:"default:0" json:"remaining"`
	Period    string         `gorm:"not null" json:"period"` // monthly, quarterly, yearly
	Year      int            `gorm:"not null" json:"year"`
	Month     int            `json:"month"`
	Status    string         `gorm:"default:active" json:"status"` // active, completed, cancelled
	CreatedBy uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Unit    SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Creator User         `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (Budget) TableName() string {
	return "budgets"
}

// FinancialReport represents generated financial reports
type FinancialReport struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID        uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Title         string         `gorm:"not null" json:"title"`
	Type          string         `gorm:"not null" json:"type"` // daily, weekly, monthly, quarterly, yearly
	PeriodStart   time.Time      `json:"periodStart"`
	PeriodEnd     time.Time      `json:"periodEnd"`
	TotalIncome   float64        `json:"totalIncome"`
	TotalExpenses float64        `json:"totalExpenses"`
	Balance       float64        `json:"balance"`
	Data          string         `gorm:"type:text" json:"data"`
	GeneratedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"generatedBy"`
	Status        string         `gorm:"default:generated" json:"status"` // generated, sent, archived
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Unit      SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Generator User         `gorm:"foreignKey:GeneratedBy" json:"generator,omitempty"`
}

func (FinancialReport) TableName() string {
	return "financial_reports"
}
