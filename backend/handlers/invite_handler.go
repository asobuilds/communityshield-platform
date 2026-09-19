package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// CreateInviteRequest is the body for POST /api/v1/invites
type CreateInviteRequest struct {
	Scope     string     `json:"scope" binding:"required"` // "platform" or "unit"
	UnitID    *uuid.UUID `json:"unitId,omitempty"`
	ExpiresIn int        `json:"expiresInHours"` // optional, hours
	MaxUses   int        `json:"maxUses"`        // optional, default 1
}

// CreateInvite creates a platform or unit invite.
// - scope=platform: any authenticated user may invite a new person to register.
// - scope=unit:     only a Unit Admin or Head Admin of a VERIFIED unit may invite.
func CreateInvite(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj := userValue.(*models.User)

	var req CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope is required"})
		return
	}

	if req.Scope != "platform" && req.Scope != "unit" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope must be 'platform' or 'unit'"})
		return
	}

	if req.Scope == "unit" {
		if req.UnitID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unitId is required for unit invites"})
			return
		}

		var membership models.UnitMembership
		if err := config.DB.
			Where("unit_id = ? AND user_id = ? AND status = ?", *req.UnitID, userObj.ID, models.MembershipActive).
			First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not an active member of this unit"})
			return
		}

		isAdmin := membership.IsHeadAdmin || membership.Role == models.UnitRoleAdmin
		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins or the head admin can create unit invites"})
			return
		}

		var unit models.SecurityUnit
		if err := config.DB.First(&unit, "id = ?", *req.UnitID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
			return
		}

		if !unit.IsVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unit must be verified before inviting members"})
			return
		}
	}

	// Generate a secure random invite code and store only its hash
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate invite code"})
		return
	}
	code := hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(code))
	codeHash := hex.EncodeToString(hash[:])

	maxUses := req.MaxUses
	if maxUses < 1 {
		maxUses = 1
	}

	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		t := time.Now().UTC().Add(time.Duration(req.ExpiresIn) * time.Hour)
		expiresAt = &t
	}

	invite := models.UnitInvite{
		UnitID:    req.UnitID,
		Scope:     req.Scope,
		CodeHash:  codeHash,
		Status:    "pending",
		ExpiresAt: expiresAt,
		MaxUses:   maxUses,
		UseCount:  0,
		CreatedBy: userObj.ID,
	}

	if err := config.DB.Create(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invite"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"invite": gin.H{
			"id":        invite.ID,
			"scope":     invite.Scope,
			"unitId":    invite.UnitID,
			"code":      code,
			"expiresAt": invite.ExpiresAt,
			"maxUses":   invite.MaxUses,
		},
		"warning": "Save this code now. It cannot be retrieved again.",
	})
}

// ValidateInviteRequest is the body for POST /api/v1/invites/validate
type ValidateInviteRequest struct {
	Code string `json:"code" binding:"required"`
}

// ValidateInvite is PUBLIC. It checks if an invite code is still valid
// and returns only the safe metadata needed to show the join prompt.
func ValidateInvite(c *gin.Context) {
	var req ValidateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	hash := sha256.Sum256([]byte(req.Code))
	codeHash := hex.EncodeToString(hash[:])

	var invite models.UnitInvite
	if err := config.DB.Where("code_hash = ?", codeHash).First(&invite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found"})
		return
	}

	if invite.Status != "pending" {
		c.JSON(http.StatusGone, gin.H{"error": "Invite is no longer valid"})
		return
	}
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now().UTC()) {
		c.JSON(http.StatusGone, gin.H{"error": "Invite has expired"})
		return
	}
	if invite.UseCount >= invite.MaxUses {
		c.JSON(http.StatusGone, gin.H{"error": "Invite has reached its maximum uses"})
		return
	}

	resp := gin.H{
		"scope": invite.Scope,
	}
	if invite.Scope == "unit" && invite.UnitID != nil {
		var unit models.SecurityUnit
		if err := config.DB.First(&unit, "id = ?", *invite.UnitID).Error; err == nil {
			resp["unitId"] = unit.ID
			resp["unitName"] = unit.Name
		}
	}

	c.JSON(http.StatusOK, gin.H{"invite": resp})
}
