package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Rating records a citizen's rating of an officer or a unit.
// One rating per (user, case, target) — enforced by unique index.
type Rating struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_rating_user_case_target" json:"userId"`
	CaseID     *uuid.UUID     `gorm:"type:uuid;index:idx_rating_case;uniqueIndex:idx_rating_user_case_target" json:"caseId,omitempty"`
	TargetID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_rating_target;uniqueIndex:idx_rating_user_case_target" json:"targetId"`
	TargetType string         `gorm:"type:varchar(16);not null;uniqueIndex:idx_rating_user_case_target" json:"targetType"` // "unit" or "officer"
	Rating     int            `gorm:"not null" json:"rating"` // 1..5
	Comment    string         `gorm:"type:text" json:"comment,omitempty"`

	Status        string     `gorm:"type:varchar(16);not null;default:'active';index:idx_rating_status" json:"status"` // active | flagged | removed
	FlaggedByID   *uuid.UUID `gorm:"type:uuid" json:"flaggedById,omitempty"`
	FlaggedReason string     `gorm:"type:text" json:"flaggedReason,omitempty"`
	FlaggedAt     *time.Time `json:"flaggedAt,omitempty"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Rating) TableName() string {
	return "ratings"
}
