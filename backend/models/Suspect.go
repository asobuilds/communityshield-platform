package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Suspect struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FirstName      string         `gorm:"not null" json:"firstName"`
	LastName       string         `gorm:"not null" json:"lastName"`
	Alias          string         `json:"alias"`
	Gender         string         `json:"gender"`
	DateOfBirth    time.Time      `json:"dateOfBirth"`
	Nationality    string         `json:"nationality"`
	IDNumber       string         `json:"idNumber"`
	Phone          string         `json:"phone"`
	Email          string         `json:"email"`
	Address        string         `json:"address"`
	Description    string         `json:"description"`
	Status         string         `gorm:"default:active" json:"status"`
	DangerLevel    string         `gorm:"default:medium" json:"dangerLevel"`
	Wanted         bool           `gorm:"default:false" json:"wanted"`
	Category       string         `gorm:"default:general" json:"category"` // theft, assault, fraud, etc.
	SubCategory    string         `json:"subCategory"`
	RiskScore      float64        `gorm:"default:0" json:"riskScore"` // AI-calculated risk score
	UnitID         *uuid.UUID     `gorm:"type:uuid" json:"unitId,omitempty"`
	TransferStatus string         `gorm:"default:''" json:"transferStatus"`
	TransferToUnit *uuid.UUID     `gorm:"type:uuid" json:"transferToUnit,omitempty"`
	TransferredBy  *uuid.UUID     `gorm:"type:uuid" json:"transferredBy,omitempty"`
	TransferReason string         `json:"transferReason"`
	PhotoURL       string         `json:"photoUrl"`
	CreatedBy      uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	CreatedByUser     User                 `gorm:"foreignKey:CreatedBy" json:"createdByUser,omitempty"`
	Unit              *SecurityUnit        `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	TransferToUnitObj *SecurityUnit        `gorm:"foreignKey:TransferToUnit" json:"transferToUnitObj,omitempty"`
	TransferredByUser *User                `gorm:"foreignKey:TransferredBy" json:"transferredByUser,omitempty"`
	Associations      []SuspectAssociation `gorm:"foreignKey:SuspectID" json:"associations,omitempty"`
}

func (Suspect) TableName() string {
	return "suspects"
}

// SuspectAssociation links suspects to cases and other suspects
type SuspectAssociation struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SuspectID       uuid.UUID      `gorm:"type:uuid;not null" json:"suspectId"`
	TargetID        uuid.UUID      `gorm:"type:uuid;not null" json:"targetId"`
	TargetType      string         `gorm:"not null" json:"targetType"`           // case, suspect
	AssociationType string         `gorm:"default:known" json:"associationType"` // known, accomplice, relative, witness
	Notes           string         `json:"notes"`
	CreatedBy       uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Suspect       Suspect `gorm:"foreignKey:SuspectID" json:"suspect,omitempty"`
	TargetCase    Case    `gorm:"foreignKey:TargetID" json:"targetCase,omitempty"`
	TargetSuspect Suspect `gorm:"foreignKey:TargetID" json:"targetSuspect,omitempty"`
	Creator       User    `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (SuspectAssociation) TableName() string {
	return "suspect_associations"
}

// SuspectSighting tracks when a suspect is seen
type SuspectSighting struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SuspectID   uuid.UUID      `gorm:"type:uuid;not null" json:"suspectId"`
	ReportedBy  uuid.UUID      `gorm:"type:uuid;not null" json:"reportedBy"`
	UnitID      uuid.UUID      `gorm:"type:uuid;not null" json:"unitId"`
	Latitude    float64        `json:"latitude"`
	Longitude   float64        `json:"longitude"`
	Location    string         `json:"location"`
	Description string         `json:"description"`
	Timestamp   time.Time      `json:"timestamp"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Suspect  Suspect      `gorm:"foreignKey:SuspectID" json:"suspect,omitempty"`
	Reporter User         `gorm:"foreignKey:ReportedBy" json:"reporter,omitempty"`
	Unit     SecurityUnit `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
}

func (SuspectSighting) TableName() string {
	return "suspect_sightings"
}

// SuspectCase links suspects to cases
type SuspectCase struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SuspectID uuid.UUID      `gorm:"type:uuid;not null" json:"suspectId"`
	CaseID    uuid.UUID      `gorm:"type:uuid;not null" json:"caseId"`
	Role      string         `json:"role"` // suspect, witness, victim, person_of_interest
	Notes     string         `json:"notes"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Suspect Suspect `gorm:"foreignKey:SuspectID" json:"suspect,omitempty"`
	Case    Case    `gorm:"foreignKey:CaseID" json:"case,omitempty"`
}

func (SuspectCase) TableName() string {
	return "suspect_cases"
}
