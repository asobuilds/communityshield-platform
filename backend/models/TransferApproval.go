package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransferType defines what is being transferred
const (
	TransferTypeSuspect = "suspect"
	TransferTypeCase    = "case"
)

// TransferApproval tracks multi-signature approvals for transfers
type TransferApproval struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TransferID   uuid.UUID      `gorm:"type:uuid;not null" json:"transferId"`
	TransferType string         `gorm:"not null" json:"transferType"` // suspect, case
	ApproverID   uuid.UUID      `gorm:"type:uuid;not null" json:"approverId"`
	UnitID       uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Status       string         `gorm:"default:pending" json:"status"` // pending, approved, rejected
	Comment      string         `json:"comment"`
	ApprovedAt   *time.Time     `json:"approvedAt,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Approver User         `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	Unit     SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
}

func (TransferApproval) TableName() string {
	return "transfer_approvals"
}

// TransferRequest represents a transfer request (unified for suspect and case)
type TransferRequest struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TargetID     uuid.UUID      `gorm:"type:uuid;not null" json:"targetId"`          // suspect_id or case_id
	TargetType   string         `gorm:"not null" json:"targetType"`                 // suspect, case
	FromUnitID   uuid.UUID      `gorm:"type:uuid;not null" json:"fromUnitId"`
	ToUnitID     uuid.UUID      `gorm:"type:uuid;not null" json:"toUnitId"`
	RequestedBy  uuid.UUID      `gorm:"type:uuid;not null" json:"requestedBy"`
	Reason       string         `json:"reason"`
	Status       string         `gorm:"default:pending" json:"status"` // pending, approved, rejected, completed
	ApprovalCount int           `gorm:"default:0" json:"approvalCount"`
	RequiredApprovals int      `gorm:"default:3" json:"requiredApprovals"`
	ApprovedAt   *time.Time     `json:"approvedAt,omitempty"`
	CompletedAt  *time.Time     `json:"completedAt,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	FromUnit   SecurityUnit `gorm:"foreignKey:FromUnitID" json:"fromUnit,omitempty"`
	ToUnit     SecurityUnit `gorm:"foreignKey:ToUnitID" json:"toUnit,omitempty"`
	RequestedByUser User    `gorm:"foreignKey:RequestedBy" json:"requestedByUser,omitempty"`
	Approvals  []TransferApproval `gorm:"foreignKey:TransferID" json:"approvals,omitempty"`
}

func (TransferRequest) TableName() string {
	return "transfer_requests"
}