package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SOSAlert struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	UnitID      *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	Latitude    float64        `json:"latitude"`
	Longitude   float64        `json:"longitude"`
	Description string         `json:"description"`
	Status      string         `gorm:"default:pending" json:"status"` // pending, dispatched, resolved, cancelled
	Priority    string         `gorm:"default:high" json:"priority"`   // high, medium, low
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Unit *SecurityUnit  `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
}

func (SOSAlert) TableName() string {
	return "sos_alerts"
}