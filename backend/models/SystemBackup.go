package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SystemBackup struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Type        string         `gorm:"default:full" json:"type"`
	Status      string         `gorm:"default:pending" json:"status"`
	FileURL     string         `json:"fileUrl"`
	Size        int64          `json:"size"`
	StartedAt   time.Time      `json:"startedAt"`
	CompletedAt *time.Time     `json:"completedAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"createdBy"`
	Creator   User      `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (SystemBackup) TableName() string {
	return "system_backups"
}