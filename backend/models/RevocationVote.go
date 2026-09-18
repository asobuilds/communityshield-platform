package models

import (
	"time"

	"github.com/google/uuid"
)

type RevocationVote struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CycleID               uuid.UUID `gorm:"type:uuid;not null;index:idx_revocation_vote_cycle;uniqueIndex:idx_revocation_vote_cycle_actor_action" json:"cycleId"`
	ActorMembershipID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_revocation_vote_cycle_actor_action" json:"actorMembershipId"`
	ActorUserID           uuid.UUID `gorm:"type:uuid;not null;index:idx_revocation_vote_actor_user" json:"actorUserId"`
	TargetMembershipID    uuid.UUID `gorm:"type:uuid;not null;index:idx_revocation_vote_target_membership" json:"targetMembershipId"`
	TargetUserID          uuid.UUID `gorm:"type:uuid;not null;index:idx_revocation_vote_target_user" json:"targetUserId"`
	Action                string    `gorm:"not null;index:idx_revocation_vote_action;uniqueIndex:idx_revocation_vote_cycle_actor_action" json:"action"`
	Choice                string    `gorm:"not null" json:"choice"`
	Reason                string    `gorm:"type:text;not null" json:"reason"`
	EligibilityStatus     string    `gorm:"not null;index:idx_revocation_vote_eligibility" json:"eligibilityStatus"`
	EligibilitySnapshot   string    `gorm:"type:text;not null" json:"eligibilitySnapshot"`
	ActorRoleAtVote       string    `gorm:"not null" json:"actorRoleAtVote"`
	TargetRoleAtVote      string    `gorm:"not null" json:"targetRoleAtVote"`
	CreatedAt             time.Time `gorm:"not null;index:idx_revocation_vote_created_at" json:"createdAt"`
}
