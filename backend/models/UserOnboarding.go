package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserOnboarding struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Step        string         `gorm:"default:welcome" json:"step"`
	Status      string         `gorm:"default:pending" json:"status"`
	CompletedAt *time.Time     `json:"completedAt"`
	Metadata    string         `gorm:"type:text" json:"metadata"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserOnboarding) TableName() string {
	return "user_onboarding"
}