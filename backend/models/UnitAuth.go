package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UnitAuth struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_unit_auth_unit" json:"unitId"`
	Rules     string         `gorm:"type:text;not null" json:"rules"`
	Version   int            `gorm:"not null;default:1" json:"version"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
