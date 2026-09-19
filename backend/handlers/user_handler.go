package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// callerIsSuperAdmin returns true if the authenticated caller is a super admin.
func callerIsSuperAdmin(c *gin.Context) (*models.User, bool) {
	value, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return nil, false
	}
	user, ok := value.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return nil, false
	}
	if !user.IsSuperAdmin && user.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
		return nil, false
	}
	return user, true
}

// GetUsers returns every user in the system. Super admin only.
func GetUsers(c *gin.Context) {
	if _, ok := callerIsSuperAdmin(c); !ok {
		return
	}

	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// UpdateUser updates another user's profile by ID. Super admin only.
// Self-service profile edits must go through /auth/profile.
func UpdateUser(c *gin.Context) {
	caller, ok := callerIsSuperAdmin(c)
	if !ok {
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Phone     string `json:"phone"`
		Status    string `json:"status"`
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

	// A super admin must not be able to suspend/demote themselves by accident.
	if user.ID == caller.ID && input.Status != "" && input.Status != user.Status {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot change your own account status"})
		return
	}

	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}
	if input.LastName != "" {
		user.LastName = input.LastName
	}
	if input.Phone != "" {
		user.Phone = input.Phone
	}
	if input.Status != "" {
		user.Status = input.Status
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    user,
	})
}

// DeleteUser deletes a user account by ID. Super admin only.
// A super admin cannot delete their own account through this endpoint.
func DeleteUser(c *gin.Context) {
	caller, ok := callerIsSuperAdmin(c)
	if !ok {
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if caller.ID == userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot delete your own account through this endpoint"})
		return
	}

	if err := config.DB.Delete(&models.User{}, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetUserProfile returns the authenticated user's own record.
func GetUserProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateUserPreferences is not yet implemented — returns 501 rather than a fake success.
func UpdateUserPreferences(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Preference persistence is not yet implemented",
	})
}