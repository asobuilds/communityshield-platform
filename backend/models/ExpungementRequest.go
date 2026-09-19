package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExpungementRequest is the audit trail + decision record for a citizen's
// request to clear their own suspect record. Only the suspect's linked citizen
// may request; only a unit head admin (or super admin) may decide.
type ExpungementRequest struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SuspectID     uuid.UUID `gorm:"type:uuid;not null;index:idx_expun_req_suspect" json:"suspectId"`

	RequestedBy   uuid.UUID `gorm:"type:uuid;not null" json:"requestedBy"` // == suspect.UserID
	RequestedAt   time.Time `gorm:"not null" json:"requestedAt"`
	Reason        string    `gorm:"type:text" json:"reason"`                // citizen's stated reason

	Status        string     `gorm:"type:varchar(16);not null;default:'pending';index:idx_expun_req_status" json:"status"` // pending | approved | denied
	ReviewedBy    *uuid.UUID `gorm:"type:uuid" json:"reviewedBy,omitempty"`
	ReviewedAt    *time.Time `json:"reviewedAt,omitempty"`
	DecisionReason string    `gorm:"type:text" json:"decisionReason,omitempty"` // approval or denial note

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Suspect  Suspect `gorm:"foreignKey:SuspectID" json:"-"`
	Reviewer User    `gorm:"foreignKey:ReviewedBy" json:"-"`
}

func (ExpungementRequest) TableName() string {
	return "expungement_requests"
}
