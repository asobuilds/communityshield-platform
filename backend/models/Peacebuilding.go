package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// PeaceCommittee represents a community peace committee
type PeaceCommittee struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID      uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Location    string         `json:"location"`
	Status      string         `gorm:"default:active" json:"status"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Unit    SecurityUnit      `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Creator User              `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Members []CommitteeMember `gorm:"foreignKey:CommitteeID" json:"members,omitempty"`
}

func (PeaceCommittee) TableName() string {
	return "peace_committees"
}

// CommitteeMember represents a member of a peace committee
type CommitteeMember struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CommitteeID uuid.UUID      `gorm:"type:uuid;not null" json:"committeeId"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"userId"`
	Role        string         `gorm:"default:member" json:"role"`
	Status      string         `gorm:"default:active" json:"status"`
	JoinedAt    time.Time      `json:"joinedAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Committee PeaceCommittee `gorm:"foreignKey:CommitteeID" json:"committee,omitempty"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (CommitteeMember) TableName() string {
	return "committee_members"
}

// ConflictResolution represents a conflict resolution case
type ConflictResolution struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID      *uuid.UUID     `gorm:"type:uuid" json:"caseId,omitempty"`
	UnitID      uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"type:text;not null" json:"description"`
	Parties     string         `gorm:"type:text" json:"parties"`
	Status      string         `gorm:"default:open" json:"status"`
	Priority    string         `gorm:"default:medium" json:"priority"`
	MediatorID  *uuid.UUID     `gorm:"type:uuid" json:"mediatorId,omitempty"`
	Resolution  string         `gorm:"type:text" json:"resolution"`
	ResolvedAt  *time.Time     `json:"resolvedAt"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Case     Case         `gorm:"foreignKey:CaseID" json:"case,omitempty"`
	Unit     SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Mediator User         `gorm:"foreignKey:MediatorID" json:"mediator,omitempty"`
	Creator  User         `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (ConflictResolution) TableName() string {
	return "conflict_resolutions"
}

// CommunityTrustScore represents trust scores for units and officers
type CommunityTrustScore struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID      *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	OfficerID   *uuid.UUID     `gorm:"type:uuid" json:"officerId,omitempty"`
	Score       float64        `gorm:"default:0" json:"score"`
	RatingCount int            `gorm:"default:0" json:"ratingCount"`
	Category    string         `gorm:"default:response_time" json:"category"`
	Trend       string         `gorm:"default:stable" json:"trend"`
	LastUpdated time.Time      `json:"lastUpdated"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Unit    SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Officer Officer      `gorm:"foreignKey:OfficerID" json:"officer,omitempty"`
}

func (CommunityTrustScore) TableName() string {
	return "community_trust_scores"
}

// PeaceMetric tracks peace metrics for communities
type PeaceMetric struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID              uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Month               string         `gorm:"not null" json:"month"`       // YYYY-MM
	PeaceLevel          float64        `gorm:"default:0" json:"peaceLevel"` // 0-100
	ConflictCount       int            `gorm:"default:0" json:"conflictCount"`
	ResolvedCount       int            `gorm:"default:0" json:"resolvedCount"`
	TrustScore          float64        `gorm:"default:0" json:"trustScore"`
	CommunityEngagement float64        `gorm:"default:0" json:"communityEngagement"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	Unit SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
}

func (PeaceMetric) TableName() string {
	return "peace_metrics"
}
