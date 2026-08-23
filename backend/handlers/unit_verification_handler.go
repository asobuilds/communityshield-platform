package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// GetUnitVerificationStatus returns the verification state of a security unit.
func GetUnitVerificationStatus(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	var unit models.SecurityUnit
	if err := config.DB.First(&unit, "id = ?", unitID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "security unit not found"})
		return
	}

	var adminCount int64
	config.DB.Model(&models.UnitMember{}).
		Where("unit_id = ? AND role = ? AND status = ?", unitID, "admin", "active").
		Count(&adminCount)

	c.JSON(http.StatusOK, gin.H{
		"unitId":               unit.ID,
		"unitName":             unit.Name,
		"isVerified":           unit.IsVerified,
		"status":               unit.Status,
		"adminCount":           adminCount,
		"minimumAdmins":        5,
		"eligibleForVerify":    adminCount >= 5,
		"verificationRequired": true,
	})
}
