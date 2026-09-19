package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// GetUnitAuth returns the parsed policy + version for a unit.
// Any verified member of the unit may read it.
func GetUnitAuth(c *gin.Context) {
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

	var membership models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ? AND verified_at IS NOT NULL", unitID, userObj.ID, models.MembershipActive).
		First(&membership).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a verified member of this unit"})
		return
	}

	svc := services.NewUnitAuthService()
	policy, version, err := svc.GetPolicy(unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policy":  policy,
		"version": version,
	})
}

// UpsertUnitAuth creates or updates the unit's policy.
// Only head admin or a unit admin may write.
func UpsertUnitAuth(c *gin.Context) {
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

	var membership models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ?", unitID, userObj.ID, models.MembershipActive).
		First(&membership).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this unit"})
		return
	}
	if !membership.IsHeadAdmin && membership.Role != models.UnitRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only head admin or admins can update unit policy"})
		return
	}

	var policy services.UnitAuthPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy payload"})
		return
	}

	svc := services.NewUnitAuthService()
	version, err := svc.UpsertPolicy(unitID, policy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "unit policy updated",
		"version": version,
	})
}

// GetPublicUnitAuth — public read-only view of a unit's governance policy.
// No authentication required. Any visitor can see the rules the unit operates under.
func GetPublicUnitAuth(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	var unit models.SecurityUnit
	if err := config.DB.First(&unit, "id = ?", unitID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
		return
	}

	svc := services.NewUnitAuthService()
	policy, version, err := svc.GetPolicy(unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unitId":   unit.ID,
		"unitName": unit.Name,
		"policy":   policy,
		"version":  version,
	})
}
