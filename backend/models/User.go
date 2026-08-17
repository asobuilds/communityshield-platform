package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email             string         `gorm:"unique;not null" json:"email"`
	Password          string         `gorm:"not null" json:"-"`
	FirstName         string         `gorm:"not null" json:"firstName"`
	LastName          string         `gorm:"not null" json:"lastName"`
	Phone             string         `json:"phone"`
	Role              string         `gorm:"default:citizen" json:"role"`
	UnitID            *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	Rank              string         `json:"rank,omitempty"`
	Status            string         `gorm:"default:active" json:"status"`
	ProfileImage      string         `json:"profileImage,omitempty"`
	ResetToken        string         `gorm:"index" json:"-"`
	ResetTokenExpires *time.Time     `gorm:"index" json:"-"`
	ReceiveEmail      bool           `gorm:"default:true" json:"receiveEmail"` // 🔥 NEW
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
        EmailVerified bool `gorm:"default:false" json:"emailVerified"`
        PhoneVerified bool `gorm:"default:false" json:"phoneVerified"`
}