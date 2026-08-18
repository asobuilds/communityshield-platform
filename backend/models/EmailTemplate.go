package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmailTemplate struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"unique;not null" json:"name"`
	Subject   string         `gorm:"not null" json:"subject"`
	Body      string         `gorm:"type:text;not null" json:"body"`
	Variables string         `gorm:"type:text" json:"variables"`
	Category  string         `json:"category"`
	IsActive  bool           `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (EmailTemplate) TableName() string {
	return "email_templates"
}