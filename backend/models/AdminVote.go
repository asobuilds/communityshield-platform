package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminVote struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ElectionID  uuid.UUID      `gorm:"type:uuid;not null;index:idx_admin_vote_election;uniqueIndex:idx_admin_vote_election_voter" json:"electionId"`
	VoterID     uuid.UUID      `gorm:"type:uuid;not null;index:idx_admin_vote_voter;uniqueIndex:idx_admin_vote_election_voter" json:"voterId"`
	CandidateID uuid.UUID      `gorm:"type:uuid;not null;index:idx_admin_vote_candidate" json:"candidateId"`
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
