package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type News struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title          string         `gorm:"not null" json:"title"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	Source         string         `json:"source"`
	Author         string         `json:"author"`
	Category       string         `gorm:"default:general" json:"category"` // general, security, community, advisory
	Location       string         `json:"location"`
	Sentiment      string         `gorm:"default:neutral" json:"sentiment"` // positive, neutral, negative
	SentimentScore float64        `gorm:"default:0" json:"sentimentScore"`
	ThreatLevel    string         `gorm:"default:low" json:"threatLevel"` // low, medium, high, critical
	Impact         string         `json:"impact"`
	Status         string         `gorm:"default:published" json:"status"` // draft, published, archived
	PublishedAt    time.Time      `json:"publishedAt"`
	CreatedBy      uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	AuthorUser User `gorm:"foreignKey:CreatedBy" json:"authorUser,omitempty"`
}

func (News) TableName() string {
	return "news"
}

// NewsAlert tracks alerts generated from news
type NewsAlert struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	NewsID    uuid.UUID      `gorm:"type:uuid;not null" json:"newsId"`
	AlertType string         `gorm:"not null" json:"alertType"` // warning, advisory, critical
	Message   string         `gorm:"type:text;not null" json:"message"`
	Location  string         `json:"location"`
	Severity  string         `gorm:"default:medium" json:"severity"`
	Status    string         `gorm:"default:active" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	News News `gorm:"foreignKey:NewsID" json:"news,omitempty"`
}

func (NewsAlert) TableName() string {
	return "news_alerts"
}
