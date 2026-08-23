package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// Camera represents a security camera
type Camera struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID    uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Name      string         `gorm:"not null" json:"name"`
	Location  string         `json:"location"`
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	RTSPURL   string         `json:"rtspUrl"` // Camera stream URL
	Status    string         `gorm:"default:active" json:"status"`
	IsActive  bool           `gorm:"default:true" json:"isActive"`
	LastPing  *time.Time     `json:"lastPing"`
	CreatedBy uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Unit    SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Creator User         `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (Camera) TableName() string {
	return "cameras"
}

// VideoAlert represents an alert from video analytics
type VideoAlert struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CameraID    uuid.UUID      `gorm:"type:uuid;not null" json:"cameraId"`
	UnitID      uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	AlertType   string         `gorm:"not null" json:"alertType"` // motion, object, person, vehicle, face
	Confidence  float64        `json:"confidence"`
	Description string         `gorm:"type:text" json:"description"`
	Severity    string         `gorm:"default:medium" json:"severity"` // low, medium, high, critical
	ImageURL    string         `json:"imageUrl"`
	VideoURL    string         `json:"videoUrl"`
	Status      string         `gorm:"default:pending" json:"status"` // pending, reviewed, resolved
	ReviewedBy  *uuid.UUID     `gorm:"type:uuid" json:"reviewedBy,omitempty"`
	ReviewedAt  *time.Time     `json:"reviewedAt,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Camera   Camera       `gorm:"foreignKey:CameraID" json:"camera,omitempty"`
	Unit     SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Reviewer User         `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
}

func (VideoAlert) TableName() string {
	return "video_alerts"
}

// SocialMediaPost represents a monitored social media post
type SocialMediaPost struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID      *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	Platform    string         `gorm:"not null" json:"platform"` // twitter, facebook, instagram, tiktok
	PostID      string         `gorm:"unique;not null" json:"postId"`
	Author      string         `gorm:"not null" json:"author"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Sentiment   string         `gorm:"default:neutral" json:"sentiment"` // positive, neutral, negative
	Confidence  float64        `json:"confidence"`
	ThreatLevel string         `gorm:"default:low" json:"threatLevel"` // low, medium, high, critical
	Location    string         `json:"location"`
	Status      string         `gorm:"default:monitored" json:"status"` // monitored, flagged, reviewed, resolved
	ReviewedBy  *uuid.UUID     `gorm:"type:uuid" json:"reviewedBy,omitempty"`
	ReviewedAt  *time.Time     `json:"reviewedAt,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Unit     SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Reviewer User         `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
}

func (SocialMediaPost) TableName() string {
	return "social_media_posts"
}
