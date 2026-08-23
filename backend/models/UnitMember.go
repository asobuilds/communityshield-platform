package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UnitMember struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	UnitID uuid.UUID `gorm:"type:uuid;not null;index" json:"unitId"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`

	Role   string `gorm:"not null;index" json:"role"`                   // admin, officer
	Status string `gorm:"not null;default:pending;index" json:"status"` // pending, active, rejected, suspended

	// Administrative verification
	Verified   bool       `gorm:"default:false;index" json:"verified"`
	VerifiedAt *time.Time `json:"verifiedAt,omitempty"`
	VerifiedBy *uuid.UUID `gorm:"type:uuid" json:"verifiedBy,omitempty"`

	// Invitation/application tracking
	InvitedBy  *uuid.UUID `gorm:"type:uuid" json:"invitedBy,omitempty"`
	ApprovedBy *uuid.UUID `gorm:"type:uuid" json:"approvedBy,omitempty"`
	ApprovedAt *time.Time `json:"approvedAt,omitempty"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Unit SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	User User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UnitMember) TableName() string {
	return "unit_members"
}
