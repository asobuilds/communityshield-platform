package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// SuperAdminMiddleware - Checks if user is super admin
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}
		userObj := user.(*models.User)

		if userObj.Role != "super_admin" && !userObj.IsSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetAllUsers - Super admin sees all users
func GetAllUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	var response []map[string]interface{}
	for _, user := range users {
		response = append(response, map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"phone":      user.Phone,
			"firstName":  user.FirstName,
			"lastName":   user.LastName,
			"role":       user.Role,
			"status":     user.Status,
			"createdAt":  user.CreatedAt,
			"lastLogin":  user.LastLogin,
			"unitId":     user.UnitID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users": response,
		"total": len(users),
	})
}

// GetUserByID - Super admin sees any user's details
func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var caseCount int64
	config.DB.Model(&models.Case{}).Where("reported_by = ?", userID).Count(&caseCount)

	var sosCount int64
	config.DB.Model(&models.SOSAlert{}).Where("user_id = ?", userID).Count(&sosCount)

	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"stats": gin.H{
			"casesReported": caseCount,
			"sosAlerts":     sosCount,
		},
	})
}

// ImpersonateUser - Super admin can login as any user
func ImpersonateUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	superAdmin, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	superAdminObj := superAdmin.(*models.User)

	var targetUser models.User
	if err := config.DB.First(&targetUser, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Log impersonation with correct AuditLog fields
	auditLog := models.AuditLog{
		UserID:     superAdminObj.ID,
		Action:     "impersonate",
		EntityType: "user",
		EntityID:   targetUser.ID.String(),
		OldValue:   targetUser.Email,
		NewValue:   "impersonated",
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Timestamp:  time.Now(),
	}
	config.DB.Create(&auditLog)

	authService := services.NewAuthService()
	token, err := authService.GenerateJWT(&targetUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	superAdminObj.Impersonating = &targetUser.ID
	config.DB.Save(superAdminObj)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Impersonating user",
		"token":          token,
		"user":           targetUser,
		"impersonating":  true,
		"originalUser":   superAdminObj.ID,
	})
}

// StopImpersonation - Stop impersonating
func StopImpersonation(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Impersonating == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not impersonating any user"})
		return
	}

	var superAdmin models.User
	if err := config.DB.First(&superAdmin, "id = ?", userObj.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Super admin not found"})
		return
	}

	superAdmin.Impersonating = nil
	config.DB.Save(&superAdmin)

	c.JSON(http.StatusOK, gin.H{
		"message": "Stopped impersonating",
		"user":    superAdmin,
	})
}

// UpdateUserRole - Super admin can change user roles
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Role = input.Role
	config.DB.Save(&user)

	// Log the change with correct AuditLog fields
	auditLog := models.AuditLog{
		UserID:     user.ID,
		Action:     "role_change",
		EntityType: "user",
		EntityID:   user.ID.String(),
		OldValue:   user.Role,
		NewValue:   input.Role,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Timestamp:  time.Now(),
	}
	config.DB.Create(&auditLog)

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated",
		"user":    user,
	})
}

// SuspendUser - Super admin can suspend users
func SuspendUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Status = "suspended"
	config.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message": "User suspended",
		"user":    user,
	})
}

// ActivateUser - Super admin can activate users
func ActivateUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Status = "active"
	config.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message": "User activated",
		"user":    user,
	})
}

// GetSystemStats - Super admin sees overall system stats
func GetSystemStats(c *gin.Context) {
	var totalUsers int64
	var totalCases int64
	var totalUnits int64
	var totalOfficers int64
	var totalSOS int64
	var totalSuspects int64

	config.DB.Model(&models.User{}).Count(&totalUsers)
	config.DB.Model(&models.Case{}).Count(&totalCases)
	config.DB.Model(&models.SecurityUnit{}).Count(&totalUnits)
	config.DB.Model(&models.Officer{}).Count(&totalOfficers)
	config.DB.Model(&models.SOSAlert{}).Count(&totalSOS)
	config.DB.Model(&models.Suspect{}).Count(&totalSuspects)

	var dailyActive int64
	config.DB.Model(&models.User{}).Where("last_login > ?", time.Now().Add(-24*time.Hour)).Count(&dailyActive)

	var pendingCases int64
	var resolvedCases int64
	config.DB.Model(&models.Case{}).Where("status = ?", "pending").Count(&pendingCases)
	config.DB.Model(&models.Case{}).Where("status = ?", "resolved").Count(&resolvedCases)

	c.JSON(http.StatusOK, gin.H{
		"totalUsers":     totalUsers,
		"totalCases":     totalCases,
		"totalUnits":     totalUnits,
		"totalOfficers":  totalOfficers,
		"totalSOS":       totalSOS,
		"totalSuspects":  totalSuspects,
		"dailyActive":    dailyActive,
		"pendingCases":   pendingCases,
		"resolvedCases":  resolvedCases,
	})
}