package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"security-solution/models"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadProfilePicture handles profile picture upload for a user
func UploadProfilePicture(c *gin.Context) {
	userIDStr := c.PostForm("userId")
	if userIDStr == "" {
		userIDStr = c.Query("userId")
	}
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}

	// Check if user exists
	var user models.User
	if err := DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("profilePicture")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profilePicture file is required"})
		return
	}

	// Validate file size (max 2MB)
	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Max size is 2MB"})
		return
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: jpg, jpeg, png, gif, webp"})
		return
	}

	// Create uploads directory if not exists
	uploadDir := "./uploads/profile-pictures"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	filename := userID.String() + ext
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Update user's ProfileImage field
	profileImageURL := "/uploads/profile-pictures/" + filename
	user.ProfileImage = profileImageURL
	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Profile picture uploaded successfully",
		"profileImage": profileImageURL,
	})
}

// GetUserProfile returns user profile (including image)
func GetUserProfile(c *gin.Context) {
	userIDStr := c.Param("userId")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}

	var user models.User
	if err := DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":           user.ID,
			"email":        user.Email,
			"firstName":    user.FirstName,
			"lastName":     user.LastName,
			"phone":        user.Phone,
			"role":         user.Role,
			"profileImage": user.ProfileImage,
			"receiveEmail": user.ReceiveEmail,
			"createdAt":    user.CreatedAt,
		},
	})
}

// UpdateUserProfile updates user's basic info (first name, last name, phone)
func UpdateUserProfile(c *gin.Context) {
	var input struct {
		UserID       string `json:"userId" binding:"required"`
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
		Phone        string `json:"phone"`
		ReceiveEmail *bool  `json:"receiveEmail"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}

	var user models.User
	if err := DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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
	if input.ReceiveEmail != nil {
		user.ReceiveEmail = *input.ReceiveEmail
	}

	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":           user.ID,
			"firstName":    user.FirstName,
			"lastName":     user.LastName,
			"phone":        user.Phone,
			"email":        user.Email,
			"role":         user.Role,
			"receiveEmail": user.ReceiveEmail,
		},
	})
}

// 🔥 NEW: GetUserUsedUnits returns units a user has reported cases to (with count)
func GetUserUsedUnits(c *gin.Context) {
	userIDStr := c.Param("userId")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}

	// Check user exists
	var user models.User
	if err := DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get cases grouped by unit
	type UnitCount struct {
		UnitID uuid.UUID
		Count  int
	}
	var results []UnitCount
	if err := DB.Model(&models.Case{}).
		Select("unit_id, count(*) as count").
		Where("reported_by = ? AND unit_id IS NOT NULL", userID).
		Group("unit_id").
		Order("count desc").
		Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch used units"})
		return
	}

	// Get unit details for each
	var usedUnits []map[string]interface{}
	for _, r := range results {
		var unit models.Unit
		if err := DB.First(&unit, "id = ?", r.UnitID).Error; err == nil {
			usedUnits = append(usedUnits, map[string]interface{}{
				"unit":  unit,
				"count": r.Count,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"usedUnits": usedUnits,
		"count":     len(usedUnits),
	})
}