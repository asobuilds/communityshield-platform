package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Evidence struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID      uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	FileURL     string         `gorm:"not null" json:"fileUrl"`
	FileType    string         `gorm:"not null" json:"fileType"`
	Description string         `json:"description"`
	UploadedBy  uuid.UUID      `gorm:"type:uuid;not null" json:"uploadedBy"`
	UploadedAt  time.Time      `gorm:"autoCreateTime" json:"uploadedAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}