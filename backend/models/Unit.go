package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SecurityUnit struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name               string         `gorm:"not null;unique" json:"name"`
	Type               string         `gorm:"not null" json:"type"`
	Latitude           float64        `json:"latitude"`
	Longitude          float64        `json:"longitude"`
	CoverageArea       string         `json:"coverageArea"`
	ContactPerson      string         `json:"contactPerson"`
	ContactPhone       string         `json:"contactPhone"`
	ContactEmail       string         `json:"contactEmail"`
	RegistrationNumber string         `gorm:"unique" json:"registrationNumber"`
	Status             string         `gorm:"default:active" json:"status"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SecurityUnit) TableName() string {
	return "security_units"
}