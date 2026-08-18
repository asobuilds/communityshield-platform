package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DataExport struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Type        string         `gorm:"not null" json:"type"`
	Format      string         `gorm:"default:json" json:"format"`
	Status      string         `gorm:"default:pending" json:"status"`
	FileURL     string         `json:"fileUrl"`
	Filters     string         `gorm:"type:text" json:"filters"`
	Size        int64          `json:"size"`
	ExpiresAt   *time.Time     `json:"expiresAt"`
	CompletedAt *time.Time     `json:"completedAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (DataExport) TableName() string {
	return "data_exports"
}