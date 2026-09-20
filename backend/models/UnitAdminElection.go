package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UnitAdminElection struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID                uuid.UUID      `gorm:"type:uuid;not null;index:idx_unit_admin_election_unit" json:"unitId"`
	ElectionType          string         `gorm:"not null;default:admin;index:idx_unit_admin_election_type" json:"electionType"`
	CycleNumber           int            `gorm:"not null;default:0;index:idx_unit_admin_election_cycle" json:"cycleNumber"`
	RotationGroup         string         `gorm:"not null;index:idx_unit_admin_election_rotation_group" json:"rotationGroup"`
	SeatCount             int            `gorm:"not null;default:0" json:"seatCount"`
	MemberCountAtElection int            `gorm:"not null;default:0" json:"memberCountAtElection"`
	EligibleVoterCount    int            `gorm:"not null;default:0" json:"eligibleVoterCount"`
	QuorumCount           int            `gorm:"not null;default:0" json:"quorumCount"`
	ExtendedOnce          bool           `gorm:"not null;default:false" json:"extendedOnce"`
	QuorumMet             bool           `gorm:"not null;default:true" json:"quorumMet"`
	TermStart             time.Time      `gorm:"not null;index:idx_unit_admin_election_term_start" json:"termStart"`
	TermEnd               time.Time      `gorm:"not null;index:idx_unit_admin_election_term_end" json:"termEnd"`
	VotingStartsAt        *time.Time     `gorm:"index:idx_unit_admin_election_voting_start" json:"votingStartsAt,omitempty"`
	VotingEndsAt          *time.Time     `gorm:"index:idx_unit_admin_election_voting_end" json:"votingEndsAt,omitempty"`
	ResultFinalizedAt     *time.Time     `gorm:"index:idx_unit_admin_election_result_at" json:"resultFinalizedAt,omitempty"`
	Status                string         `gorm:"not null;default:draft;index:idx_unit_admin_election_status" json:"status"`
	CreatedBy             uuid.UUID      `gorm:"type:uuid;not null;index:idx_unit_admin_election_created_by" json:"createdBy"`
	CreatedAt             time.Time      `json:"createdAt"`
	UpdatedAt             time.Time      `json:"updatedAt"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}
