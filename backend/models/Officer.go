package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Officer struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID      uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_officer_unit_officer_id" json:"unitId"`
	Name        string         `gorm:"not null" json:"name"`
	Rank        string         `gorm:"not null" json:"rank"`
	BadgeNumber string         `gorm:"unique;not null" json:"badgeNumber"`
	UnitOfficerID string       `gorm:"not null;uniqueIndex:idx_officer_unit_officer_id" json:"unitOfficerId"`
	Role        string         `gorm:"not null" json:"role"`
	Phone       string         `json:"phone"`
	Email       string         `json:"email"`
	JoinedDate  time.Time      `gorm:"default:now()" json:"joinedDate"`
	Status      string         `gorm:"default:active" json:"status"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Unit SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
}

func (Officer) TableName() string {
	return "officers"
}
