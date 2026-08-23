package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Case struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID          uuid.UUID  `gorm:"type:uuid;not null" json:"unitId"`
	ReportedBy      uuid.UUID  `gorm:"type:uuid;not null" json:"reportedBy"`
	AssignedTo      *uuid.UUID `gorm:"type:uuid" json:"assignedTo,omitempty"`
	Title           string     `gorm:"not null" json:"title"`
	Description     string     `gorm:"type:text;not null" json:"description"`
	IncidentDate    time.Time  `json:"incidentDate"`
	Location        string     `json:"location"`
	Latitude        float64    `json:"latitude"`
	Longitude       float64    `json:"longitude"`
	Status          string     `gorm:"default:pending" json:"status"`
	Priority        string     `gorm:"default:medium" json:"priority"`
	TransferDetails string     `json:"transferDetails,omitempty"`
	IsPublic        bool       `gorm:"default:true" json:"isPublic"`

	// NEW FIELDS FOR AUTOMATION
	TrackingID    string     `gorm:"unique;not null" json:"trackingId"`
	PriorityLevel string     `gorm:"default:P3" json:"priorityLevel"`
	GISLatitude   float64    `json:"gisLatitude"`
	GISLongitude  float64    `json:"gisLongitude"`
	AssignedAt    *time.Time `json:"assignedAt,omitempty"`
	DispatchedAt  *time.Time `json:"dispatchedAt,omitempty"`
	ArrivedAt     *time.Time `json:"arrivedAt,omitempty"`
	ClosedAt      *time.Time `json:"closedAt,omitempty"`
	ClosedBy      *uuid.UUID `gorm:"type:uuid" json:"closedBy,omitempty"`
	ApprovedBy    *uuid.UUID `gorm:"type:uuid" json:"approvedBy,omitempty"`
	FinalReport   string     `gorm:"type:text" json:"finalReport"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Evidence []Evidence `gorm:"foreignKey:CaseID" json:"evidence,omitempty"`
	Progress []Progress `gorm:"foreignKey:CaseID" json:"progress,omitempty"`
}

func (Case) TableName() string {
	return "cases"
}

type Evidence struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID      uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	UploadedBy  uuid.UUID      `gorm:"type:uuid;not null" json:"uploadedBy"`
	Type        string         `gorm:"not null" json:"type"`
	FileURL     string         `gorm:"not null" json:"fileUrl"`
	Description string         `json:"description"`
	Latitude    float64        `json:"latitude"`
	Longitude   float64        `json:"longitude"`
	IsVerified  bool           `gorm:"default:false" json:"isVerified"`
	UploadedAt  time.Time      `json:"uploadedAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Evidence) TableName() string {
	return "evidence"
}

type Progress struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID      uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	OfficerID   uuid.UUID      `gorm:"type:uuid;not null" json:"officerId"`
	Action      string         `gorm:"not null" json:"action"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Progress) TableName() string {
	return "case_progress"
}
