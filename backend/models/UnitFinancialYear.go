package models

import (
	"time"

	"github.com/google/uuid"
)

// UnitFinancialYear is a frozen snapshot of a unit's finances for one year.
// Rows are written once when the year closes and never updated.
type UnitFinancialYear struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID          uuid.UUID `gorm:"type:uuid;not null;index:idx_fin_year_unit" json:"unitId"`
	Year            int       `gorm:"not null;uniqueIndex:idx_fin_year_unit_year" json:"year"`
	TotalIn         float64   `gorm:"not null;default:0" json:"totalIn"`
	TotalOut        float64   `gorm:"not null;default:0" json:"totalOut"`
	Net             float64   `gorm:"not null;default:0" json:"net"`
	InCount         int64     `gorm:"not null;default:0" json:"inCount"`
	OutCount        int64     `gorm:"not null;default:0" json:"outCount"`
	TopCategories   string    `gorm:"type:text" json:"topCategories,omitempty"`
	Status          string    `gorm:"type:varchar(16);not null;default:'closed'" json:"status"`
	ClosedAt        time.Time `gorm:"not null" json:"closedAt"`
	ClosedBy        uuid.UUID `gorm:"type:uuid;not null" json:"closedBy"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (UnitFinancialYear) TableName() string {
	return "unit_financial_years"
}
