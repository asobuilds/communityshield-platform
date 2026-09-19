package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// OpenRevocationRequest — body for POST /units/:unitId/revocations
type OpenRevocationRequest struct {
	TargetMembershipID uuid.UUID `json:"targetMembershipId" binding:"required"`
	CycleType          string    `json:"cycleType" binding:"required"` // regular_admin or head_admin
	Reason             string    `json:"reason" binding:"required"`
}

// OpenRevocationCycle opens a new revocation cycle for a target membership.
func OpenRevocationCycle(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj := userValue.(*models.User)

	var initiator models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ? AND verified_at IS NOT NULL", unitID, userObj.ID, models.MembershipActive).
		First(&initiator).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a verified member of this unit"})
		return
	}

	var req OpenRevocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "targetMembershipId, cycleType, reason are required"})
		return
	}

	svc := services.NewRevocationService()
	cycle, err := svc.OpenRevocationCycle(unitID, req.TargetMembershipID, initiator.ID, req.CycleType, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"cycle": cycle})
}

// CastRevocationVoteRequest — body for POST /revocations/:id/vote
type CastRevocationVoteRequest struct {
	Action string `json:"action" binding:"required"` // member_vote or head_admin_approval
	Choice string `json:"choice" binding:"required"` // support / oppose / abstain / approve / reject
	Reason string `json:"reason" binding:"required"`
}

// CastRevocationVote records one vote or approval on a cycle.
func CastRevocationVote(c *gin.Context) {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj := userValue.(*models.User)

	var req CastRevocationVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action, choice, reason are required"})
		return
	}

	var actor models.UnitMembership
	if err := config.DB.
		Where("user_id = ? AND status = ?", userObj.ID, models.MembershipActive).
		First(&actor).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not an active member"})
		return
	}

	svc := services.NewRevocationService()
	if err := svc.CastVote(cycleID, actor.ID, userObj.ID, req.Action, req.Choice, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vote recorded"})
}

// CloseRevocationCycleRequest — body for POST /revocations/:id/close
type CloseRevocationCycleRequest struct {
	HeadAdminApprovalMembershipID *uuid.UUID `json:"headAdminApprovalMembershipId,omitempty"`
}

// CloseRevocationCycle finalizes a cycle and applies removal if thresholds pass.
func CloseRevocationCycle(c *gin.Context) {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj := userValue.(*models.User)

	var cycle models.RevocationCycle
	if err := config.DB.First(&cycle, "id = ?", cycleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cycle not found"})
		return
	}

	var membership models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ?", cycle.UnitID, userObj.ID, models.MembershipActive).
		First(&membership).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this unit"})
		return
	}
	if !membership.IsHeadAdmin && membership.Role != models.UnitRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only head admin or admins can close revocation cycles"})
		return
	}

	var req CloseRevocationCycleRequest
	_ = c.ShouldBindJSON(&req)

	svc := services.NewRevocationService()
	if err := svc.CloseRevocationCycle(cycleID, req.HeadAdminApprovalMembershipID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "revocation cycle closed"})
}

// GetRevocationCycle returns a cycle and its votes (admins and head admin only).
func GetRevocationCycle(c *gin.Context) {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj := userValue.(*models.User)

	var cycle models.RevocationCycle
	if err := config.DB.First(&cycle, "id = ?", cycleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cycle not found"})
		return
	}

	var membership models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ?", cycle.UnitID, userObj.ID, models.MembershipActive).
		First(&membership).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this unit"})
		return
	}
	if !membership.IsHeadAdmin && membership.Role != models.UnitRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only head admin or admins can view cycle details"})
		return
	}

	var votes []models.RevocationVote
	config.DB.Where("cycle_id = ?", cycleID).Order("created_at ASC").Find(&votes)

	c.JSON(http.StatusOK, gin.H{
		"cycle": cycle,
		"votes": votes,
	})
}