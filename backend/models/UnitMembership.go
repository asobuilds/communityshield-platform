package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

const (
    UnitRoleAdmin   = "unit_admin"
    UnitRoleOfficer = "officer"
)

const (
    MembershipPending = "pending"
    MembershipActive  = "active"
    MembershipRejected = "rejected"
    MembershipRevoked = "revoked"
)

type UnitMembership struct {
    ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

    UnitID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_unit_user" json:"unitId"`
    UserID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_unit_user" json:"userId"`

    Role   string `gorm:"not null;index" json:"role"`
    Status string `gorm:"not null;default:pending;index" json:"status"`

    InvitedBy *uuid.UUID `gorm:"type:uuid" json:"invitedBy,omitempty"`

    AcceptedAt *time.Time `json:"acceptedAt,omitempty"`
    RevokedAt  *time.Time `json:"revokedAt,omitempty"`

    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    Unit UserUnit `gorm:"foreignKey:UnitID" json:"-"`
    User User     `gorm:"foreignKey:UserID" json:"-"`
}

func (UnitMembership) TableName() string {
    return "unit_memberships"
}

// UserUnit is intentionally a lightweight relationship type used by
// UnitMembership without changing the existing SecurityUnit model.
type UserUnit struct {
    ID uuid.UUID `gorm:"type:uuid;primary_key"`
}

func (UserUnit) TableName() string {
    return "security_units"
}
