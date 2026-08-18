package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/models"
)

// GetPublicSettings gets all public settings
func GetPublicSettings(c *gin.Context) {
	var settings []models.SystemSettings
	if err := config.DB.Where("is_public = ?", true).Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
	})
}

// GetSystemSetting gets a system setting
func GetSystemSetting(c *gin.Context) {
	key := c.Param("key")

	var setting models.SystemSettings
	if err := config.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"setting": setting,
	})
}

// UpdateSystemSetting updates a system setting (admin only)
func UpdateSystemSetting(c *gin.Context) {
	key := c.Param("key")

	var input struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
		IsPublic    bool   `json:"isPublic"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update settings"})
		return
	}

	var setting models.SystemSettings
	if err := config.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		setting = models.SystemSettings{
			Key:         key,
			Value:       input.Value,
			Description: input.Description,
			Category:    "general",
			IsPublic:    input.IsPublic,
			UpdatedBy:   &userObj.ID,
		}
		if err := config.DB.Create(&setting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create setting"})
			return
		}
	} else {
		setting.Value = input.Value
		if input.Description != "" {
			setting.Description = input.Description
		}
		setting.IsPublic = input.IsPublic
		setting.UpdatedBy = &userObj.ID
		if err := config.DB.Save(&setting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting updated successfully",
		"setting": setting,
	})
}

// GetEmailTemplate gets an email template
func GetEmailTemplate(c *gin.Context) {
	name := c.Param("name")

	var template models.EmailTemplate
	if err := config.DB.Where("name = ?", name).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"template": template,
	})
}

// UpdateEmailTemplate updates an email template (admin only)
func UpdateEmailTemplate(c *gin.Context) {
	name := c.Param("name")

	var input struct {
		Subject   string `json:"subject" binding:"required"`
		Body      string `json:"body" binding:"required"`
		Variables string `json:"variables"`
		Category  string `json:"category"`
		IsActive  bool   `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update email templates"})
		return
	}

	var template models.EmailTemplate
	if err := config.DB.Where("name = ?", name).First(&template).Error; err != nil {
		template = models.EmailTemplate{
			Name:      name,
			Subject:   input.Subject,
			Body:      input.Body,
			Variables: input.Variables,
			Category:  input.Category,
			IsActive:  input.IsActive,
		}
		if err := config.DB.Create(&template).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
			return
		}
	} else {
		template.Subject = input.Subject
		template.Body = input.Body
		if input.Variables != "" {
			template.Variables = input.Variables
		}
		if input.Category != "" {
			template.Category = input.Category
		}
		template.IsActive = input.IsActive
		if err := config.DB.Save(&template).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Template updated successfully",
		"template": template,
	})
}

// CreateDataExport creates a data export request
func CreateDataExport(c *gin.Context) {
	var input struct {
		Type    string `json:"type" binding:"required"`
		Format  string `json:"format"`
		Filters string `json:"filters"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if input.Format == "" {
		input.Format = "json"
	}

	export := models.DataExport{
		UserID:  userObj.ID,
		Type:    input.Type,
		Format:  input.Format,
		Filters: input.Filters,
		Status:  "pending",
	}

	if err := config.DB.Create(&export).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create export"})
		return
	}

	now := time.Now()
	export.Status = "completed"
	export.CompletedAt = &now
	export.FileURL = "/exports/" + export.ID.String() + "." + input.Format
	config.DB.Save(&export)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Export created successfully",
		"export":  export,
	})
}

// GetDataExports gets all data exports for a user
func GetDataExports(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var exports []models.DataExport
	if err := config.DB.Where("user_id = ?", userObj.ID).Order("created_at desc").Find(&exports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exports": exports,
	})
}

// GetUserOnboarding gets user onboarding status
func GetUserOnboarding(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var onboarding models.UserOnboarding
	if err := config.DB.Where("user_id = ?", userObj.ID).First(&onboarding).Error; err != nil {
		onboarding = models.UserOnboarding{
			UserID: userObj.ID,
			Step:   "welcome",
			Status: "pending",
		}
		config.DB.Create(&onboarding)
	}

	c.JSON(http.StatusOK, gin.H{
		"onboarding": onboarding,
	})
}

// UpdateUserOnboarding updates user onboarding progress
func UpdateUserOnboarding(c *gin.Context) {
	var input struct {
		Step   string `json:"step" binding:"required"`
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var onboarding models.UserOnboarding
	if err := config.DB.Where("user_id = ?", userObj.ID).First(&onboarding).Error; err != nil {
		onboarding = models.UserOnboarding{
			UserID: userObj.ID,
			Step:   input.Step,
			Status: input.Status,
		}
	} else {
		onboarding.Step = input.Step
		onboarding.Status = input.Status
		if input.Status == "completed" {
			now := time.Now()
			onboarding.CompletedAt = &now
		}
	}

	if err := config.DB.Save(&onboarding).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update onboarding"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Onboarding updated",
		"onboarding": onboarding,
	})
}