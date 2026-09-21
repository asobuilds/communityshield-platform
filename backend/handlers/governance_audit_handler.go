package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// GetGovernanceAudit returns a summary of governance activity for a unit.
// Head admin of the unit, or super admin, may call.
func GetGovernanceAudit(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj, ok := userValue.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	if !userObj.IsSuperAdmin && userObj.Role != "super_admin" {
		var membership models.UnitMembership
		if err := config.DB.
			Where("unit_id = ? AND user_id = ? AND status = ? AND is_head_admin = ?",
				unitID, userObj.ID, models.MembershipActive, true).
			First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the head admin of this unit may view governance audit"})
			return
		}
	}

	// Elections summary
	var totalElections int64
	var finalized int64
	var lowTurnout int64
	var totalVotes int64
	config.DB.Model(&models.UnitAdminElection{}).
		Where("unit_id = ?", unitID).Count(&totalElections)
	config.DB.Model(&models.UnitAdminElection{}).
		Where("unit_id = ? AND status IN ?", unitID, []string{"finalized", "finalized_low_turnout"}).
		Count(&finalized)
	config.DB.Model(&models.UnitAdminElection{}).
		Where("unit_id = ? AND status = ?", unitID, "finalized_low_turnout").
		Count(&lowTurnout)
	config.DB.Model(&models.AdminVote{}).
		Joins("JOIN unit_admin_elections ON unit_admin_elections.id = admin_votes.election_id").
		Where("unit_admin_elections.unit_id = ?", unitID).
		Count(&totalVotes)

	// Revocation summary
	var totalCycles int64
	var completedCycles int64
	var rejectedCycles int64
	config.DB.Model(&models.RevocationCycle{}).
		Where("unit_id = ?", unitID).Count(&totalCycles)
	config.DB.Model(&models.RevocationCycle{}).
		Where("unit_id = ? AND status = ?", unitID, "completed").
		Count(&completedCycles)
	config.DB.Model(&models.RevocationCycle{}).
		Where("unit_id = ? AND status = ?", unitID, "rejected").
		Count(&rejectedCycles)

	// Appeals summary
	var totalAppeals int64
	var upheldAppeals int64
	var overturnedAppeals int64
	config.DB.Model(&models.Appeal{}).
		Joins("JOIN revocation_cycles ON revocation_cycles.id = appeals.revocation_cycle_id").
		Where("revocation_cycles.unit_id = ?", unitID).
		Count(&totalAppeals)
	config.DB.Model(&models.Appeal{}).
		Joins("JOIN revocation_cycles ON revocation_cycles.id = appeals.revocation_cycle_id").
		Where("revocation_cycles.unit_id = ? AND appeals.status = ?", unitID, "upheld").
		Count(&upheldAppeals)
	config.DB.Model(&models.Appeal{}).
		Joins("JOIN revocation_cycles ON revocation_cycles.id = appeals.revocation_cycle_id").
		Where("revocation_cycles.unit_id = ? AND appeals.status = ?", unitID, "overturned").
		Count(&overturnedAppeals)

	// Current admins + head admin
	var adminCount int64
	var headAdminCount int64
	config.DB.Model(&models.UnitMembership{}).
		Where("unit_id = ? AND status = ? AND role = ?", unitID, models.MembershipActive, models.UnitRoleAdmin).
		Count(&adminCount)
	config.DB.Model(&models.UnitMembership{}).
		Where("unit_id = ? AND status = ? AND is_head_admin = ?", unitID, models.MembershipActive, true).
		Count(&headAdminCount)

	// Recent elections (last 5) — brief summary only
	type electionBrief struct {
		ID            uuid.UUID `json:"id"`
		Status        string    `json:"status"`
		SeatCount     int       `json:"seatCount"`
		QuorumCount   int       `json:"quorumCount"`
		ExtendedOnce  bool      `json:"extendedOnce"`
		QuorumMet     bool      `json:"quorumMet"`
		ElectionType  string    `json:"electionType"`
		TermStart     string    `json:"termStart"`
		TermEnd       string    `json:"termEnd"`
	}
	var elections []models.UnitAdminElection
	config.DB.Where("unit_id = ?", unitID).
		Order("created_at DESC").Limit(5).Find(&elections)

	briefs := make([]electionBrief, 0, len(elections))
	for _, e := range elections {
		briefs = append(briefs, electionBrief{
			ID:           e.ID,
			Status:       e.Status,
			SeatCount:    e.SeatCount,
			QuorumCount:  e.QuorumCount,
			ExtendedOnce: e.ExtendedOnce,
			QuorumMet:    e.QuorumMet,
			ElectionType: e.ElectionType,
			TermStart:    e.TermStart.Format("2006-01-02"),
			TermEnd:      e.TermEnd.Format("2006-01-02"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"unitId": unitID,
		"elections": gin.H{
			"total":      totalElections,
			"finalized":  finalized,
			"lowTurnout": lowTurnout,
			"totalVotes": totalVotes,
			"recent":     briefs,
		},
		"revocations": gin.H{
			"total":     totalCycles,
			"completed": completedCycles,
			"rejected":  rejectedCycles,
		},
		"appeals": gin.H{
			"total":      totalAppeals,
			"upheld":     upheldAppeals,
			"overturned": overturnedAppeals,
		},
		"currentAdmins": gin.H{
			"adminCount":     adminCount,
			"headAdminCount": headAdminCount,
		},
	})
}
