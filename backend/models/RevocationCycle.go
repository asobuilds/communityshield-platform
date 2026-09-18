package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RevocationCycle struct {
	ID                                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID                              uuid.UUID      `gorm:"type:uuid;not null;index:idx_revocation_cycle_unit_target" json:"unitId"`
	TargetMembershipID                  uuid.UUID      `gorm:"type:uuid;not null;index:idx_revocation_cycle_unit_target" json:"targetMembershipId"`
	TargetRole                          string         `gorm:"not null" json:"targetRole"`
	TargetMembershipStatus              string         `gorm:"not null" json:"targetMembershipStatus"`
	CycleType                           string         `gorm:"not null;index:idx_revocation_cycle_type" json:"cycleType"`
	Status                              string         `gorm:"not null;default:open;index:idx_revocation_cycle_status" json:"status"`
	EligibleMemberCount                 int            `gorm:"not null;default:0" json:"eligibleMemberCount"`
	EligibilitySnapshotAt               time.Time      `gorm:"not null;index:idx_revocation_cycle_eligibility_snapshot_at" json:"eligibilitySnapshotAt"`
	QuorumRequired                      int            `gorm:"not null;default:0" json:"quorumRequired"`
	RequiredVotes                       int            `gorm:"not null;default:0" json:"requiredVotes"`
	ForVotes                            int            `gorm:"not null;default:0" json:"forVotes"`
	AgainstVotes                        int            `gorm:"not null;default:0" json:"againstVotes"`
	AbstainVotes                        int            `gorm:"not null;default:0" json:"abstainVotes"`
	TotalVotes                          int            `gorm:"not null;default:0" json:"totalVotes"`
	DistinctEligibleVoters              int            `gorm:"not null;default:0" json:"distinctEligibleVoters"`
	HeadAdminApproved                   bool           `gorm:"not null;default:false" json:"headAdminApproved"`
	HeadAdminApprovalActorMembershipID  *uuid.UUID     `gorm:"type:uuid" json:"headAdminApprovalActorMembershipId,omitempty"`
	HeadAdminApprovalReason             *string        `gorm:"type:text" json:"headAdminApprovalReason,omitempty"`
	HeadAdminApprovedAt                 *time.Time     `json:"headAdminApprovedAt,omitempty"`
	InitiatedByMembershipID             uuid.UUID      `gorm:"type:uuid;not null;index:idx_revocation_cycle_initiator" json:"initiatedByMembershipId"`
	Reason                              string         `gorm:"type:text;not null" json:"reason"`
	UnitAuthPolicyVersion               string         `gorm:"not null;index:idx_revocation_cycle_unit_auth_policy_version" json:"unitAuthPolicyVersion"`
	OpenedAt                            time.Time      `gorm:"not null;index:idx_revocation_cycle_opened_at" json:"openedAt"`
	ClosedAt                            *time.Time     `json:"closedAt,omitempty"`
	CompletedAt                         *time.Time     `json:"completedAt,omitempty"`
	CoolingOffUntil                     *time.Time     `json:"coolingOffUntil,omitempty"`
	CreatedAt                           time.Time      `json:"createdAt"`
	UpdatedAt                           time.Time      `json:"updatedAt"`
	DeletedAt                           gorm.DeletedAt `gorm:"index" json:"-"`
}
