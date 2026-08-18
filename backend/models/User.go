package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email     string         `gorm:"unique;not null" json:"email"`
	Phone     string         `gorm:"unique" json:"phone"`
	FirstName string         `gorm:"not null" json:"firstName"`
	LastName  string         `gorm:"not null" json:"lastName"`
	Password  string         `gorm:"not null" json:"-"`
	Role      string         `gorm:"default:citizen" json:"role"`
	UnitID    *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	Status    string         `gorm:"default:pending" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}