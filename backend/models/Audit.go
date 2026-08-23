package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type AuditLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Action     string         `gorm:"not null" json:"action"`
	EntityType string         `gorm:"not null" json:"entityType"`
	EntityID   string         `gorm:"not null" json:"entityId"`
	OldValue   string         `gorm:"type:text" json:"oldValue"`
	NewValue   string         `gorm:"type:text" json:"newValue"`
	IPAddress  string         `json:"ipAddress"`
	UserAgent  string         `json:"userAgent"`
	Timestamp  time.Time      `json:"timestamp"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type SystemHealth struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CPUUsage       float64        `json:"cpuUsage"`
	MemoryUsage    float64        `json:"memoryUsage"`
	DiskUsage      float64        `json:"diskUsage"`
	ActiveUsers    int            `json:"activeUsers"`
	TotalRequests  int            `json:"totalRequests"`
	ResponseTime   float64        `json:"responseTime"`
	DatabaseStatus string         `gorm:"default:healthy" json:"databaseStatus"`
	ServerStatus   string         `gorm:"default:healthy" json:"serverStatus"`
	Uptime         int64          `json:"uptime"`
	LastCheck      time.Time      `json:"lastCheck"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SystemHealth) TableName() string {
	return "system_health"
}

type ActivityLog struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	SessionID    string         `json:"sessionId"`
	ActivityType string         `gorm:"not null" json:"activityType"`
	Description  string         `json:"description"`
	IPAddress    string         `json:"ipAddress"`
	Device       string         `json:"device"`
	Location     string         `json:"location"`
	Duration     int            `json:"duration"`
	CreatedAt    time.Time      `json:"createdAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}

type NotificationLog struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Type         string         `gorm:"not null" json:"type"`
	Title        string         `gorm:"not null" json:"title"`
	Message      string         `gorm:"type:text;not null" json:"message"`
	Status       string         `gorm:"default:sent" json:"status"`
	SentAt       time.Time      `json:"sentAt"`
	ReadAt       *time.Time     `json:"readAt"`
	DeliveryData string         `gorm:"type:text" json:"deliveryData"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (NotificationLog) TableName() string {
	return "notification_logs"
}
