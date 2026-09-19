package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/models"
	"security-solution/services"
)

// AgeOverride allows a super admin to verify DOB, grant/revoke a minor
// exception, or correct a date of birth. Every action is audited.
// PUT /api/v1/users/:id/age-override
func AgeOverride(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	actorValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	actorObj := actorValue.(*models.User)

	var input struct {
		Decision    string `json:"decision" binding:"required"`             // verify_dob | grant_minor_exception | revoke_minor_exception | dob_corrected
		DateOfBirth string `json:"dateOfBirth"`                               // optional; required for dob_corrected
		Reason      string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := services.NewAgeService()
	user, audit, err := svc.ApplyAgeOverride(services.AgeOverrideInput{
		UserID:      userID,
		ActorID:     actorObj.ID,
		Decision:    input.Decision,
		DateOfBirth: input.DateOfBirth,
		Reason:      input.Reason,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "invalid decision") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":                   user.ID,
			"email":                user.Email,
			"minorStatus":          user.MinorStatus,
			"dobVerified":          user.DOBVerified,
			"minorExceptionGranted": user.MinorExceptionGranted,
		},
		"audit": audit,
	})
}
