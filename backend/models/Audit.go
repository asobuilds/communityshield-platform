package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog tracks all system activities
type AuditLog struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      *uuid.UUID     `gorm:"type:uuid" json:"userId,omitempty"`
	UnitID      *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	Action      string         `gorm:"not null" json:"action"`
	Resource    string         `gorm:"not null" json:"resource"`
	ResourceID  string         `json:"resourceId"`
	Details     string         `gorm:"type:text" json:"details"`
	IPAddress   string         `json:"ipAddress"`
	UserAgent   string         `json:"userAgent"`
	Status      string         `gorm:"default:success" json:"status"` // success, failure, warning
	Severity    string         `gorm:"default:info" json:"severity"`   // info, warning, error, critical
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Unit SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// SystemHealth tracks system health metrics
type SystemHealth struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CPUUsage         float64        `json:"cpuUsage"`
	MemoryUsage      float64        `json:"memoryUsage"`
	DiskUsage        float64        `json:"diskUsage"`
	ActiveUsers      int            `json:"activeUsers"`
	TotalRequests    int            `json:"totalRequests"`
	ResponseTime     float64        `json:"responseTime"`
	DatabaseStatus   string         `gorm:"default:healthy" json:"databaseStatus"`
	ServerStatus     string         `gorm:"default:healthy" json:"serverStatus"`
	Uptime           int64          `json:"uptime"` // in seconds
	LastCheck        time.Time      `json:"lastCheck"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SystemHealth) TableName() string {
	return "system_health"
}

// ActivityLog tracks user activity sessions
type ActivityLog struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	SessionID   string         `json:"sessionId"`
	ActivityType string        `gorm:"not null" json:"activityType"` // login, logout, view, create, update, delete
	Description string         `json:"description"`
	IPAddress   string         `json:"ipAddress"`
	Device      string         `json:"device"`
	Location    string         `json:"location"`
	Duration    int            `json:"duration"` // in seconds
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}

// NotificationLog tracks notifications sent
type NotificationLog struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Type         string         `gorm:"not null" json:"type"` // email, sms, push, in-app
	Title        string         `gorm:"not null" json:"title"`
	Message      string         `gorm:"type:text;not null" json:"message"`
	Status       string         `gorm:"default:sent" json:"status"` // sent, delivered, read, failed
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