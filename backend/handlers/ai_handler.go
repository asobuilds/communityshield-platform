package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// AIAnalyzeLocation analyzes security risk for a location
func AIAnalyzeLocation(c *gin.Context) {
	var input struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		Location  string  `json:"location"`
		Radius    float64 `json:"radius"`
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

	var incidents []models.Case
	query := config.DB.Model(&models.Case{}).Where("status != ?", "closed")

	if userObj.Role == "officer" || userObj.Role == "unit_admin" {
		if userObj.UnitID != nil {
			query = query.Where("unit_id = ?", userObj.UnitID)
		}
	}

	if err := query.Order("created_at desc").Limit(20).Find(&incidents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
		return
	}

	recentIncidents := ""
	for i, inc := range incidents {
		if i >= 10 {
			break
		}
		recentIncidents += fmt.Sprintf("Title: %s\nStatus: %s\nLocation: %s\n---\n", inc.Title, inc.Status, inc.Location)
	}

	aiService := services.NewAIService()
	analysis, err := aiService.AnalyzeLocationRisk(input.Latitude, input.Longitude, input.Location, recentIncidents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"location":      input.Location,
		"latitude":      input.Latitude,
		"longitude":     input.Longitude,
		"analysis":      analysis,
		"incidentCount": len(incidents),
	})
}

// AIGetMapInsights provides insights for map markers
func AIGetMapInsights(c *gin.Context) {
	var input struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
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

	var cases []models.Case
	query := config.DB.Model(&models.Case{}).Where("status != ?", "closed")

	if userObj.Role == "officer" || userObj.Role == "unit_admin" {
		if userObj.UnitID != nil {
			query = query.Where("unit_id = ?", userObj.UnitID)
		}
	}

	if err := query.Order("created_at desc").Limit(30).Find(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases"})
		return
	}

	var incidents []string
	for _, c := range cases {
		incidents = append(incidents, fmt.Sprintf("Title: %s | Location: %s | Status: %s", c.Title, c.Location, c.Status))
	}

	aiService := services.NewAIService()
	analysis, err := aiService.AnalyzeIncidentPatterns("Current View", incidents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis":   analysis,
		"caseCount":  len(cases),
		"viewLat":    input.Latitude,
		"viewLng":    input.Longitude,
	})
}

// AIGenerateSecurityWarning generates security warnings
func AIGenerateSecurityWarning(c *gin.Context) {
	var input struct {
		Location    string   `json:"location"`
		IncidentIDs []string `json:"incidentIds"`
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can generate security warnings"})
		return
	}

	var incidentsData string
	if len(input.IncidentIDs) > 0 {
		var cases []models.Case
		if err := config.DB.Where("id IN (?)", input.IncidentIDs).Find(&cases).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
			return
		}
		for _, c := range cases {
			incidentsData += fmt.Sprintf("Title: %s\nDescription: %s\nLocation: %s\nStatus: %s\n---\n", c.Title, c.Description, c.Location, c.Status)
		}
	} else {
		var cases []models.Case
		query := config.DB.Model(&models.Case{}).Where("status != ?", "closed")
		if input.Location != "" {
			query = query.Where("location ILIKE ?", "%"+input.Location+"%")
		}
		if userObj.UnitID != nil {
			query = query.Where("unit_id = ?", userObj.UnitID)
		}
		if err := query.Order("created_at desc").Limit(10).Find(&cases).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
			return
		}
		for _, c := range cases {
			incidentsData += fmt.Sprintf("Title: %s\nLocation: %s\nStatus: %s\n---\n", c.Title, c.Location, c.Status)
		}
	}

	if incidentsData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No incident data found to generate warning"})
		return
	}

	aiService := services.NewAIService()
	warning, err := aiService.GenerateSecurityWarning(incidentsData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate warning: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"warning":     warning,
		"location":    input.Location,
		"generatedAt": time.Now(),
	})
}

// AIAnalyzeNews analyzes news for security implications
func AIAnalyzeNews(c *gin.Context) {
	var input struct {
		Content string `json:"content" binding:"required"`
		Source  string `json:"source"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	aiService := services.NewAIService()
	analysis, err := aiService.AnalyzeNewsSentiment(input.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "News analysis failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis":   analysis,
		"source":     input.Source,
		"analyzedAt": time.Now(),
	})
}

// AIGetSmartTips provides context-aware safety tips
func AIGetSmartTips(c *gin.Context) {
	var input struct {
		Location      string `json:"location"`
		TimeOfDay     string `json:"timeOfDay"`
		RecentThreats string `json:"recentThreats"`
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

	if input.TimeOfDay == "" {
		hour := time.Now().Hour()
		if hour < 6 {
			input.TimeOfDay = "Night (Late)"
		} else if hour < 12 {
			input.TimeOfDay = "Morning"
		} else if hour < 18 {
			input.TimeOfDay = "Afternoon"
		} else {
			input.TimeOfDay = "Evening"
		}
	}

	var recentIncidents []models.Case
	query := config.DB.Model(&models.Case{}).Where("created_at > ?", time.Now().Add(-7*24*time.Hour))
	if userObj.UnitID != nil {
		query = query.Where("unit_id = ?", userObj.UnitID)
	}
	if input.Location != "" {
		query = query.Where("location ILIKE ?", "%"+input.Location+"%")
	}
	query.Order("created_at desc").Limit(5).Find(&recentIncidents)

	recentThreats := input.RecentThreats
	if recentThreats == "" && len(recentIncidents) > 0 {
		for _, inc := range recentIncidents {
			recentThreats += inc.Title + ", "
		}
	}

	aiService := services.NewAIService()
	tips, err := aiService.GetSmartSafetyTips(input.Location, userObj.Role, input.TimeOfDay, recentThreats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get safety tips: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tips":       tips,
		"location":   input.Location,
		"timeOfDay":  input.TimeOfDay,
		"userRole":   userObj.Role,
	})
}

// AIPredictHotspots predicts risk hotspots
func AIPredictHotspots(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can predict hotspots"})
		return
	}

	var cases []models.Case
	query := config.DB.Model(&models.Case{}).Where("created_at > ?", time.Now().Add(-30*24*time.Hour))
	if userObj.UnitID != nil {
		query = query.Where("unit_id = ?", userObj.UnitID)
	}

	if err := query.Find(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases"})
		return
	}

	if len(cases) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Not enough data for prediction",
		})
		return
	}

	historicalData := ""
	for _, c := range cases {
		historicalData += fmt.Sprintf("Title: %s\nLocation: %s\nStatus: %s\n---\n", c.Title, c.Location, c.Status)
	}

	aiService := services.NewAIService()
	prediction, err := aiService.PredictRiskHotspots(historicalData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Prediction failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prediction":  prediction,
		"caseCount":   len(cases),
		"generatedAt": time.Now(),
	})
}

// AIGenerateCommunityAlert generates community alerts
func AIGenerateCommunityAlert(c *gin.Context) {
	var input struct {
		AlertType string `json:"alertType" binding:"required"`
		Location  string `json:"location" binding:"required"`
		Data      string `json:"data" binding:"required"`
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can generate community alerts"})
		return
	}

	aiService := services.NewAIService()
	alertContent, err := aiService.GenerateCommunityAlert(input.AlertType, input.Location, input.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate alert: " + err.Error()})
		return
	}

	alertObj := models.Alert{
		Title:     input.AlertType + " Alert - " + input.Location,
		Content:   alertContent,
		Type:      input.AlertType,
		Location:  input.Location,
		Severity:  "high",
		Status:    "active",
		CreatedBy: userObj.ID,
	}

	if err := config.DB.Create(&alertObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alert":       alertContent,
		"alertId":     alertObj.ID,
		"location":    input.Location,
		"generatedAt": time.Now(),
	})
}