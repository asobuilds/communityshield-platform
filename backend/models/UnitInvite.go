package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UnitInvite struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID        *uuid.UUID     `gorm:"type:uuid;index:idx_unit_invite_unit" json:"unitId,omitempty"`
	Scope         string         `gorm:"not null;default:unit;index:idx_unit_invite_scope" json:"scope"`
	CodeHash      string         `gorm:"type:varchar(64);unique;not null;index:idx_unit_invite_code_hash" json:"-"`
	Status        string         `gorm:"not null;default:pending;index:idx_unit_invite_status" json:"status"`
	ExpiresAt     *time.Time     `gorm:"index:idx_unit_invite_expires_at" json:"expiresAt,omitempty"`
	MaxUses       int            `gorm:"default:1" json:"maxUses"`
	UseCount      int            `gorm:"default:0" json:"useCount"`
	RevokedAt     *time.Time     `json:"revokedAt,omitempty"`
	RevokedReason *string        `gorm:"type:text" json:"revokedReason,omitempty"`
	CreatedBy     uuid.UUID      `gorm:"type:uuid;not null;index:idx_unit_invite_created_by" json:"createdBy"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
