package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseTransfer struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID        uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	RequestedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"requestedBy"`
	FromOfficerID *uuid.UUID     `gorm:"type:uuid" json:"fromOfficerId,omitempty"`
	ToOfficerID   *uuid.UUID     `gorm:"type:uuid" json:"toOfficerId,omitempty"`
	FromUnitID    *uuid.UUID     `gorm:"type:uuid" json:"fromUnitId,omitempty"`
	ToUnitID      *uuid.UUID     `gorm:"type:uuid" json:"toUnitId,omitempty"`
	Reason        string         `gorm:"type:text" json:"reason"`
	Status        string         `gorm:"default:pending" json:"status"` // pending, approved, rejected, completed
	ApprovedBy    *uuid.UUID     `gorm:"type:uuid" json:"approvedBy,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}