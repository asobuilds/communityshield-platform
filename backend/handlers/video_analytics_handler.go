package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// AddCamera adds a camera to a unit
func AddCamera(c *gin.Context) {
	var input struct {
		UnitID    string  `json:"unitId" binding:"required"`
		Name      string  `json:"name" binding:"required"`
		Location  string  `json:"location"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		RTSPURL   string  `json:"rtspUrl"`
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can add cameras"})
		return
	}

	unitID, err := uuid.Parse(input.UnitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	camera := models.Camera{
		UnitID:    unitID,
		Name:      input.Name,
		Location:  input.Location,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
		RTSPURL:   input.RTSPURL,
		Status:    "active",
		IsActive:  true,
		CreatedBy: userObj.ID,
	}

	if err := config.DB.Create(&camera).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add camera"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Camera added successfully",
		"camera":  camera,
	})
}

// GetCameras gets all cameras for a unit
func GetCameras(c *gin.Context) {
	unitID := c.Param("unitId")
	id, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view cameras"})
		return
	}

	var cameras []models.Camera
	if err := config.DB.Where("unit_id = ?", id).Find(&cameras).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cameras"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cameras": cameras,
	})
}

// GenerateVideoAlert generates an alert from video analysis
func GenerateVideoAlert(c *gin.Context) {
	var input struct {
		CameraID    string  `json:"cameraId" binding:"required"`
		AlertType   string  `json:"alertType" binding:"required"`
		Confidence  float64 `json:"confidence"`
		Description string  `json:"description"`
		Severity    string  `json:"severity"`
		ImageURL    string  `json:"imageUrl"`
		VideoURL    string  `json:"videoUrl"`
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can generate alerts"})
		return
	}

	cameraID, err := uuid.Parse(input.CameraID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid camera ID"})
		return
	}

	var camera models.Camera
	if err := config.DB.First(&camera, "id = ?", cameraID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Camera not found"})
		return
	}

	if input.Severity == "" {
		input.Severity = "medium"
	}

	alert := models.VideoAlert{
		CameraID:    cameraID,
		UnitID:      camera.UnitID,
		AlertType:   input.AlertType,
		Confidence:  input.Confidence,
		Description: input.Description,
		Severity:    input.Severity,
		ImageURL:    input.ImageURL,
		VideoURL:    input.VideoURL,
		Status:      "pending",
	}

	if err := config.DB.Create(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate alert"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Alert generated successfully",
		"alert":   alert,
	})
}

// GetVideoAlerts gets all video alerts
func GetVideoAlerts(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var alerts []models.VideoAlert
	query := config.DB.Preload("Camera").Preload("Unit").Order("created_at desc")

	if userObj.Role != "super_admin" {
		if userObj.UnitID != nil {
			query = query.Where("unit_id = ?", userObj.UnitID)
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view alerts"})
			return
		}
	}

	if err := query.Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
	})
}

// ReviewVideoAlert reviews a video alert
func ReviewVideoAlert(c *gin.Context) {
	id := c.Param("id")
	alertID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	var input struct {
		Status  string `json:"status" binding:"required"`
		Comment string `json:"comment"`
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can review alerts"})
		return
	}

	var alert models.VideoAlert
	if err := config.DB.First(&alert, "id = ?", alertID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	now := time.Now()
	alert.Status = input.Status
	alert.ReviewedBy = &userObj.ID
	alert.ReviewedAt = &now

	if err := config.DB.Save(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to review alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert reviewed",
		"alert":   alert,
	})
}

// MonitorSocialMedia monitors social media posts
func MonitorSocialMedia(c *gin.Context) {
	var input struct {
		Platform  string `json:"platform" binding:"required"`
		PostID    string `json:"postId" binding:"required"`
		Author    string `json:"author" binding:"required"`
		Content   string `json:"content" binding:"required"`
		Sentiment string `json:"sentiment"`
		Location  string `json:"location"`
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can monitor social media"})
		return
	}

	// Analyze sentiment if not provided
	if input.Sentiment == "" {
		input.Sentiment = analyzeSentiment(input.Content)
	}

	post := models.SocialMediaPost{
		UnitID:      userObj.UnitID,
		Platform:    input.Platform,
		PostID:      input.PostID,
		Author:      input.Author,
		Content:     input.Content,
		Sentiment:   input.Sentiment,
		Confidence:  0.85,
		ThreatLevel: determineThreatLevel(input.Content),
		Location:    input.Location,
		Status:      "monitored",
	}

	if err := config.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to monitor post"})
		return
	}

	// If threat level is high or critical, create alert
	if post.ThreatLevel == "high" || post.ThreatLevel == "critical" {
		// Create notification for admins
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Post monitored successfully",
		"post":    post,
	})
}

// GetSocialMediaPosts gets all monitored posts
func GetSocialMediaPosts(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var posts []models.SocialMediaPost
	query := config.DB.Order("created_at desc")

	if userObj.Role != "super_admin" {
		if userObj.UnitID != nil {
			query = query.Where("unit_id = ? OR unit_id IS NULL", userObj.UnitID)
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view posts"})
			return
		}
	}

	if err := query.Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

// Helper functions
func analyzeSentiment(content string) string {
	// Simple sentiment analysis - in production, use OpenRouter or NLP library
	negativeWords := []string{"threat", "attack", "danger", "emergency", "violence", "kill", "hurt", "destroy"}
	positiveWords := []string{"safe", "secure", "help", "support", "protect", "rescue"}

	negativeCount := 0
	positiveCount := 0

	for _, word := range negativeWords {
		if containsWord(content, word) {
			negativeCount++
		}
	}
	for _, word := range positiveWords {
		if containsWord(content, word) {
			positiveCount++
		}
	}

	if negativeCount > positiveCount {
		return "negative"
	}
	if positiveCount > negativeCount {
		return "positive"
	}
	return "neutral"
}

func determineThreatLevel(content string) string {
	criticalWords := []string{"emergency", "critical", "immediate", "urgent", "attack", "kill"}
	highWords := []string{"threat", "danger", "violence", "weapon"}

	for _, word := range criticalWords {
		if containsWord(content, word) {
			return "critical"
		}
	}
	for _, word := range highWords {
		if containsWord(content, word) {
			return "high"
		}
	}
	return "low"
}

func containsWord(content, word string) bool {
	// Simple contains check - in production, use proper word matching
	return len(content) >= len(word) && (content == word || len(content) > len(word) && containsSubstring(content, word))
}
