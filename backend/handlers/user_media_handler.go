package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// UploadAvatar accepts a multipart image, stores it, updates the user's avatar path.
func UploadAvatar(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	storageSvc := services.NewFileStorageService()
	stored, err := storageSvc.Save(file, "avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldPath := userObj.AvatarPath

	if err := config.DB.Model(&models.User{}).
		Where("id = ?", userObj.ID).
		Update("avatar_path", stored.RelativePath).Error; err != nil {
		_ = storageSvc.Delete(stored.RelativePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar"})
		return
	}

	if oldPath != "" && oldPath != stored.RelativePath {
		_ = storageSvc.Delete(oldPath)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Avatar uploaded",
		"avatarPath": stored.RelativePath,
		"hash":       stored.Hash,
		"size":       stored.Size,
	})
}

// UploadCover accepts a multipart image, stores it, updates the user's cover path.
func UploadCover(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	storageSvc := services.NewFileStorageService()
	stored, err := storageSvc.Save(file, "cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldPath := userObj.CoverPath

	if err := config.DB.Model(&models.User{}).
		Where("id = ?", userObj.ID).
		Update("cover_path", stored.RelativePath).Error; err != nil {
		_ = storageSvc.Delete(stored.RelativePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cover"})
		return
	}

	if oldPath != "" && oldPath != stored.RelativePath {
		_ = storageSvc.Delete(oldPath)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cover uploaded",
		"coverPath": stored.RelativePath,
		"hash":      stored.Hash,
		"size":      stored.Size,
	})
}

// DeleteAvatar clears the user's avatar path and removes the file.
func DeleteAvatar(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	if userObj.AvatarPath != "" {
		_ = services.NewFileStorageService().Delete(userObj.AvatarPath)
	}

	if err := config.DB.Model(&models.User{}).
		Where("id = ?", userObj.ID).
		Update("avatar_path", "").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Avatar removed"})
}

// DeleteCover clears the user's cover path and removes the file.
func DeleteCover(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	if userObj.CoverPath != "" {
		_ = services.NewFileStorageService().Delete(userObj.CoverPath)
	}

	if err := config.DB.Model(&models.User{}).
		Where("id = ?", userObj.ID).
		Update("cover_path", "").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cover"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cover removed"})
}
