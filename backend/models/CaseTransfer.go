package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type CaseTransfer struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID        uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	FromUnitID    uuid.UUID      `gorm:"type:uuid;not null" json:"fromUnitId"`
	ToUnitID      uuid.UUID      `gorm:"type:uuid;not null" json:"toUnitId"`
	TransferredBy uuid.UUID      `gorm:"type:uuid;not null" json:"transferredBy"`
	TransferNotes string         `json:"transferNotes"`
	Status        string         `gorm:"default:pending" json:"status"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Case              Case         `gorm:"foreignKey:CaseID" json:"case,omitempty"`
	FromUnit          SecurityUnit `gorm:"foreignKey:FromUnitID" json:"fromUnit,omitempty"`
	ToUnit            SecurityUnit `gorm:"foreignKey:ToUnitID" json:"toUnit,omitempty"`
	TransferredByUser User         `gorm:"foreignKey:TransferredBy" json:"transferredByUser,omitempty"`
}

func (CaseTransfer) TableName() string {
	return "case_transfers"
}
