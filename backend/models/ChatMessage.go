package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type ChatMessage struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SenderID  uuid.UUID      `gorm:"type:uuid;not null" json:"senderId"`
	CaseID    *uuid.UUID     `gorm:"type:uuid" json:"caseId,omitempty"`
	UnitID    *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Type      string         `gorm:"default:chat" json:"type"` // chat, walkie
	Room      string         `gorm:"not null" json:"room"`     // case:{id} or unit:{id}
	Timestamp time.Time      `gorm:"not null" json:"timestamp"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
