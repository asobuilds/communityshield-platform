package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Unit struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name               string         `gorm:"not null;unique" json:"name"`
	Type               string         `gorm:"not null" json:"type"` // vigilante, neighborhood_watch, community_police
	Latitude           float64        `json:"latitude"`
	Longitude          float64        `json:"longitude"`
	CoverageArea       string         `json:"coverageArea"`
	ContactPerson      string         `json:"contactPerson"`
	ContactPhone       string         `json:"contactPhone"`
	ContactEmail       string         `json:"contactEmail"`
	RegistrationNumber string         `gorm:"unique" json:"registrationNumber"` // Government authorisation
	Motto              string         `json:"motto"`                            // Unit motto
	ProfileImage       string         `json:"profileImage,omitempty"`           // Unit profile picture
	Status             string         `gorm:"default:active" json:"status"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}