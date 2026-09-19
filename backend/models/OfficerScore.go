package models

import (
	"time"

	"github.com/google/uuid"
)

// OfficerScore is a cached Bayesian aggregate of an officer's ratings.
// Recomputed by the scheduler. Read-only from handlers.
type OfficerScore struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OfficerID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_officer_score_officer" json:"officerId"`
	UnitID          uuid.UUID `gorm:"type:uuid;not null;index:idx_officer_score_unit" json:"unitId"`
	BayesianScore   float64   `gorm:"not null;default:0" json:"bayesianScore"`
	RatingCount     int       `gorm:"not null;default:0" json:"ratingCount"`
	Tier            string    `gorm:"type:varchar(16);not null;default:'unranked';index:idx_officer_score_tier" json:"tier"`
	LastComputedAt  time.Time `gorm:"not null" json:"lastComputedAt"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (OfficerScore) TableName() string {
	return "officer_scores"
}
