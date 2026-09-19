package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Appeal is a removed admin/officer's request to overturn a completed
// revocation cycle. One appeal per revocation cycle (enforced by unique index).
// Status is the single source of truth: pending | upheld | overturned.
type Appeal struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RevocationCycleID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_appeal_cycle;index:idx_appeal_cycle_status" json:"revocationCycleId"`

	AppellantUserID uuid.UUID `gorm:"type:uuid;not null;index:idx_appeal_appellant" json:"appellantUserId"`
	FiledAt         time.Time `gorm:"not null;index:idx_appeal_filed_at" json:"filedAt"`
	Reason          string    `gorm:"type:text" json:"reason,omitempty"`

	Status        string     `gorm:"type:varchar(16);not null;default:'pending';index:idx_appeal_status" json:"status"`        // pending | upheld | overturned
	DecisionReason string    `gorm:"type:text" json:"decisionReason,omitempty"`
	DecidedBy     *uuid.UUID `gorm:"type:uuid" json:"decidedBy,omitempty"`
	DecidedAt     *time.Time `json:"decidedAt,omitempty"`

	EscalatedAt   *time.Time `json:"escalatedAt,omitempty"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	RevocationCycle RevocationCycle `gorm:"foreignKey:RevocationCycleID" json:"-"`
	Appellant       User            `gorm:"foreignKey:AppellantUserID" json:"-"`
	Decider         *User           `gorm:"foreignKey:DecidedBy" json:"-"`
}

func (Appeal) TableName() string {
	return "appeals"
}
