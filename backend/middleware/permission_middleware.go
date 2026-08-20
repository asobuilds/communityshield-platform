package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// RequireRole checks if user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}
		userObj := user.(*models.User)

		for _, role := range roles {
			if userObj.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

// CanAccessCase checks if user can access a specific case
func CanAccessCase(c *gin.Context) {
	caseID := c.Param("id")
	id, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		c.Abort()
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		c.Abort()
		return
	}
	userObj := user.(*models.User)

	// Super Admin can access everything
	if userObj.Role == "super_admin" {
		c.Next()
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		c.Abort()
		return
	}

	// Citizen: only their own cases
	if userObj.Role == "citizen" {
		if caseObj.ReportedBy != userObj.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own cases"})
			c.Abort()
			return
		}
		c.Next()
		return
	}

	// Officer: cases in their unit
	if userObj.Role == "officer" {
		if userObj.UnitID == nil || caseObj.UnitID != *userObj.UnitID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only access cases in your unit"})
			c.Abort()
			return
		}
		c.Next()
		return
	}

	// Unit Admin: cases in their unit
	if userObj.Role == "unit_admin" {
		if userObj.UnitID == nil || caseObj.UnitID != *userObj.UnitID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only access cases in your unit"})
			c.Abort()
			return
		}
		c.Next()
		return
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
	c.Abort()
}

// CanAccessUnit checks if user can access a specific unit
func CanAccessUnit(c *gin.Context) {
	unitID := c.Param("unitId")
	id, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		c.Abort()
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		c.Abort()
		return
	}
	userObj := user.(*models.User)

	if userObj.Role == "super_admin" {
		c.Next()
		return
	}

	if userObj.UnitID == nil || *userObj.UnitID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this unit"})
		c.Abort()
		return
	}

	c.Next()
}