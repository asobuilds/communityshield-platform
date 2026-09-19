package models

import (
	"time"

	"github.com/google/uuid"
)

// UnitScore is a cached composite score for a unit. Recomputed by the scheduler.
// Composite = 50% officer aggregate + 30% direct unit ratings + 20% case velocity.
type UnitScore struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID             uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_unit_score_unit" json:"unitId"`

	OfficerAvg         float64   `gorm:"not null;default:0" json:"officerAvg"`
	OfficerCount       int       `gorm:"not null;default:0" json:"officerCount"`
	BayesianOfficer    float64   `gorm:"not null;default:0" json:"bayesianOfficer"`

	DirectAvg          float64   `gorm:"not null;default:0" json:"directAvg"`
	DirectCount        int       `gorm:"not null;default:0" json:"directCount"`
	BayesianDirect     float64   `gorm:"not null;default:0" json:"bayesianDirect"`

	VelocityRatio      float64   `gorm:"not null;default:0" json:"velocityRatio"` // 0..1

	CompositeScore     float64   `gorm:"not null;default:0;index:idx_unit_score_composite" json:"compositeScore"`
	Tier               string    `gorm:"type:varchar(16);not null;default:'unranked';index:idx_unit_score_tier" json:"tier"`

	LastComputedAt     time.Time `gorm:"not null" json:"lastComputedAt"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func (UnitScore) TableName() string {
	return "unit_scores"
}
