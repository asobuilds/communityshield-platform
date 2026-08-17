package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIAnalysis struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SourceType  string         `gorm:"not null" json:"sourceType"` // case, announcement, rss, social
	SourceID    string         `json:"sourceId"`
	OriginalText string        `gorm:"type:text" json:"originalText"`
	Summary     string         `gorm:"type:text" json:"summary"`
	Sentiment   string         `json:"sentiment"` // positive, neutral, negative
	ThreatLevel string         `json:"threatLevel"` // low, medium, high, critical
	Keywords    string         `json:"keywords"`
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}