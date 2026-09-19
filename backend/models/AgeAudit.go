package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgeAudit is an append-only record of every super-admin change to a user's
// age-gating state (DOB verification, minor exception, or DOB correction).
type AgeAudit struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;index:idx_age_audit_user" json:"userId"`
	ActorID            uuid.UUID      `gorm:"type:uuid;not null;index:idx_age_audit_actor" json:"actorId"`
	Action             string         `gorm:"type:varchar(32);not null;index:idx_age_audit_action" json:"action"`            // verify_dob | grant_minor_exception | revoke_minor_exception | dob_corrected
	Reason             string         `gorm:"type:text;not null" json:"reason"`
	OldMinorStatus     string         `gorm:"type:varchar(16)" json:"oldMinorStatus,omitempty"`
	NewMinorStatus     string         `gorm:"type:varchar(16)" json:"newMinorStatus,omitempty"`
	OldMinorException  bool           `gorm:"not null;default:false" json:"oldMinorException"`
	NewMinorException  bool           `gorm:"not null;default:false" json:"newMinorException"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	User  User `gorm:"foreignKey:UserID" json:"-"`
	Actor User `gorm:"foreignKey:ActorID" json:"-"`
}

func (AgeAudit) TableName() string {
	return "age_audit"
}
