package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// isHeadAdminOfAnyUnit returns true if the user is head admin of at least one unit.
func isHeadAdminOfAnyUnit(user *models.User) bool {
	if user == nil {
		return false
	}
	var count int64
	config.DB.Model(&models.UnitMembership{}).
		Where("user_id = ? AND status = ? AND is_head_admin = ?",
			user.ID, models.MembershipActive, true).
		Count(&count)
	return count > 0
}

// canManageNews returns true if the caller may create or manage news.
// Restricted to super admins and head admins. The AI system writes news
// through internal service calls, not this HTTP handler, so no role is
// required for the automated path.
func canManageNews(user *models.User) bool {
	if user == nil {
		return false
	}
	if user.IsSuperAdmin || user.Role == "super_admin" {
		return true
	}
	return isHeadAdminOfAnyUnit(user)
}

// CreateNews creates news with AI sentiment analysis
func CreateNews(c *gin.Context) {
	var input struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Source   string `json:"source"`
		Author   string `json:"author"`
		Category string `json:"category"`
		Location string `json:"location"`
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

	if !canManageNews(userObj) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only super admins and head admins can create news"})
		return
	}

	// Analyze news sentiment with AI
	aiService := services.NewAIService()
	sentimentAnalysis, err := aiService.AnalyzeNewsSentiment(input.Content)
	if err != nil {
		// Fallback to default if AI fails
		sentimentAnalysis = "Sentiment analysis not available"
	}

	// Extract sentiment from analysis (simplified)
	sentiment := "neutral"
	sentimentScore := 0.0
	threatLevel := "low"

	// Simple heuristic based on sentiment text
	if len(sentimentAnalysis) > 0 {
		if contains(sentimentAnalysis, "negative") || contains(sentimentAnalysis, "threat") || contains(sentimentAnalysis, "danger") {
			sentiment = "negative"
			sentimentScore = -0.5
			threatLevel = "medium"
		}
		if contains(sentimentAnalysis, "critical") || contains(sentimentAnalysis, "extreme") || contains(sentimentAnalysis, "emergency") {
			sentiment = "negative"
			sentimentScore = -1.0
			threatLevel = "critical"
		}
		if contains(sentimentAnalysis, "positive") || contains(sentimentAnalysis, "safe") || contains(sentimentAnalysis, "resolved") {
			sentiment = "positive"
			sentimentScore = 0.5
			threatLevel = "low"
		}
	}

	news := models.News{
		Title:          input.Title,
		Content:        input.Content,
		Source:         input.Source,
		Author:         input.Author,
		Category:       input.Category,
		Location:       input.Location,
		Sentiment:      sentiment,
		SentimentScore: sentimentScore,
		ThreatLevel:    threatLevel,
		Status:         "published",
		PublishedAt:    time.Now(),
		CreatedBy:      userObj.ID,
	}

	if news.Category == "" {
		news.Category = "general"
	}

	if err := config.DB.Create(&news).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create news"})
		return
	}

	// Generate alert if threat level is high or critical
	if threatLevel == "medium" || threatLevel == "critical" {
		alertMessage := fmt.Sprintf("News Alert: %s - Threat Level: %s\n\n%s\n\nSource: %s",
			news.Title, threatLevel, news.Content[:min(len(news.Content), 200)], news.Source)

		alert := models.NewsAlert{
			NewsID:    news.ID,
			AlertType: "warning",
			Message:   alertMessage,
			Location:  news.Location,
			Severity:  threatLevel,
			Status:    "active",
		}
		config.DB.Create(&alert)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":           "News created successfully",
		"news":              news,
		"sentimentAnalysis": sentimentAnalysis,
	})
}

// GetAllNews gets all news with role-based filtering
func GetAllNews(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var news []models.News
	query := config.DB.Model(&models.News{}).Where("status = ?", "published")

	// Role-based filtering
	if userObj.Role == "citizen" {
		// Citizens see only low-threat, general news
		query = query.Where("threat_level IN (?) AND category IN (?)",
			[]string{"low", "medium"},
			[]string{"general", "community"})
	} else if userObj.Role == "officer" {
		// Officers see all except critical (admins handle those)
		query = query.Where("threat_level != ?", "critical")
	} else if userObj.Role == "unit_admin" {
		// Admins see all news
		// No additional filter
	}

	// Filter by location if unit admin or officer
	if (userObj.Role == "unit_admin" || userObj.Role == "officer") && userObj.UnitID != nil {
		var unit models.SecurityUnit
		if err := config.DB.First(&unit, "id = ?", userObj.UnitID).Error; err == nil {
			query = query.Where("location = ? OR location = '' OR location IS NULL", unit.Name)
		}
	}

	if err := query.Order("published_at desc").Find(&news).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch news"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"news": news,
	})
}

// GetNewsByID gets specific news
func GetNewsByID(c *gin.Context) {
	id := c.Param("id")
	newsID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid news ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var news models.News
	if err := config.DB.First(&news, "id = ?", newsID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
		return
	}

	// Role-based access check
	if userObj.Role == "citizen" && (news.ThreatLevel == "critical" || news.ThreatLevel == "high") {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this news"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"news": news,
	})
}

// AIGenerateNewsSummary generates AI summary of news
func AIGenerateNewsSummary(c *gin.Context) {
	var input struct {
		NewsID string `json:"newsId" binding:"required"`
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

	if !canManageNews(userObj) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only super admins and head admins can generate AI news summaries"})
		return
	}

	newsID, err := uuid.Parse(input.NewsID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid news ID"})
		return
	}

	var news models.News
	if err := config.DB.First(&news, "id = ?", newsID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
		return
	}

	aiService := services.NewAIService()
	summary, err := aiService.SummarizeCase(news.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI summary failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"newsId":  news.ID,
		"summary": summary,
	})
}

// AIGetNewsInsights gets AI insights for news
func AIGetNewsInsights(c *gin.Context) {
	var input struct {
		NewsID string `json:"newsId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	newsID, err := uuid.Parse(input.NewsID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid news ID"})
		return
	}

	var news models.News
	if err := config.DB.First(&news, "id = ?", newsID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
		return
	}

	aiService := services.NewAIService()
	insights, err := aiService.AnalyzeNewsSentiment(news.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI insights failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"newsId":   news.ID,
		"insights": insights,
	})
}

// GetNewsAlerts gets news alerts
func GetNewsAlerts(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var alerts []models.NewsAlert
	query := config.DB.Preload("News").Where("status = ?", "active")

	if userObj.Role == "citizen" {
		// Citizens see only low and medium severity alerts
		query = query.Where("severity IN (?)", []string{"low", "medium"})
	}

	if err := query.Order("created_at desc").Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
	})
}

// helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
