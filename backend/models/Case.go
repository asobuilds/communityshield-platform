package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Case struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID      *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	ReportedBy  uuid.UUID      `gorm:"type:uuid;not null" json:"reportedBy"`
	AssignedTo  *uuid.UUID     `gorm:"type:uuid" json:"assignedTo,omitempty"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"type:text;not null" json:"description"`
	IncidentDate time.Time     `json:"incidentDate"`
	Location    string         `json:"location"`
	Latitude    float64        `json:"latitude"`
	Longitude   float64        `json:"longitude"`
	Status      string         `gorm:"default:pending" json:"status"`
	Priority    string         `gorm:"default:medium" json:"priority"`
	TransferDetails string     `json:"transferDetails,omitempty"`
	IsPublic    bool           `gorm:"default:true" json:"isPublic"`
	IsAnonymous bool           `gorm:"default:false" json:"isAnonymous"`
	Archived    bool           `gorm:"default:false;index" json:"archived"` // 🔥 NEW
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Evidence []Evidence `gorm:"foreignKey:CaseID" json:"evidence,omitempty"`
}