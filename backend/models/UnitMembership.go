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

    JoinedViaInvite  bool       `gorm:"default:false" json:"joinedViaInvite"`
    IsHeadAdmin      bool       `gorm:"default:false;index:idx_unit_membership_is_head_admin" json:"isHeadAdmin"`
    VoteCount        int        `gorm:"default:0" json:"voteCount"`
    RevokedReason    *string    `gorm:"type:text" json:"revokedReason,omitempty"`
    VerifiedAt       *time.Time `gorm:"index:idx_unit_membership_verified_at" json:"verifiedAt,omitempty"`
    ElectedAt        *time.Time `gorm:"index:idx_unit_membership_elected_at" json:"electedAt,omitempty"`
    TermStartAt      *time.Time `gorm:"index:idx_unit_membership_term_start_at" json:"termStartAt,omitempty"`
    TermEndAt        *time.Time `gorm:"index:idx_unit_membership_term_end_at" json:"termEndAt,omitempty"`
    ConsecutiveTerms int        `gorm:"default:0" json:"consecutiveTerms"`
    CoolingOffUntil  *time.Time `gorm:"index:idx_unit_membership_cooling_off_until" json:"coolingOffUntil,omitempty"`

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
