package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"security-solution/models"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
)

// callOpenRouter sends a prompt to OpenRouter and returns the response
func callOpenRouter(prompt string) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY not set")
	}

	url := "https://openrouter.ai/api/v1/chat/completions"

	payload := map[string]interface{}{
		"model": "deepseek/deepseek-r1:free", // Free model
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no response from OpenRouter")
	}

	msg, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	content, ok := msg["content"].(string)
	if !ok {
		return "", fmt.Errorf("no content in response")
	}

	return content, nil
}

// SummarizeText uses OpenRouter to summarize any text
func SummarizeText(c *gin.Context) {
	var input struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prompt := fmt.Sprintf(`Summarize the following text concisely (max 200 words). Extract key points, threat level (low/medium/high/critical if security-related), and sentiment (positive/neutral/negative):

%s`, input.Text)

	summary, err := callOpenRouter(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI failed: " + err.Error()})
		return
	}

	// Save analysis to DB
	analysis := models.AIAnalysis{
		SourceType:   "text",
		OriginalText: input.Text,
		Summary:      summary,
		Sentiment:    "neutral",
		ThreatLevel:  "low",
		Keywords:     "",
	}
	DB.Create(&analysis)

	c.JSON(http.StatusOK, gin.H{
		"summary":   summary,
		"analysis":  analysis,
	})
}

// SummarizeCase generates AI summary for a case
func SummarizeCase(c *gin.Context) {
	caseID := c.Param("id")
	parsedID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var caseItem models.Case
	if err := DB.First(&caseItem, "id = ?", parsedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	text := fmt.Sprintf("Case: %s\nDescription: %s\nLocation: %s\nStatus: %s",
		caseItem.Title, caseItem.Description, caseItem.Location, caseItem.Status)

	prompt := fmt.Sprintf(`Analyze this security case and provide:
1. A brief summary (max 100 words)
2. Threat level (low/medium/high/critical)
3. Recommended action

Text: %s`, text)

	summary, err := callOpenRouter(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI failed: " + err.Error()})
		return
	}

	// Extract threat level from response
	threatLevel := "medium"
	if strings.Contains(strings.ToLower(summary), "critical") {
		threatLevel = "critical"
	} else if strings.Contains(strings.ToLower(summary), "high") {
		threatLevel = "high"
	} else if strings.Contains(strings.ToLower(summary), "low") {
		threatLevel = "low"
	}

	analysis := models.AIAnalysis{
		SourceType:   "case",
		SourceID:     caseID,
		OriginalText: text,
		Summary:      summary,
		Sentiment:    "neutral",
		ThreatLevel:  threatLevel,
		Keywords:     "",
	}
	DB.Create(&analysis)

	c.JSON(http.StatusOK, gin.H{
		"summary":     summary,
		"threatLevel": threatLevel,
		"analysis":    analysis,
	})
}

// MonitorRSSFeeds scrapes RSS feeds and creates announcements for security threats
func MonitorRSSFeeds(c *gin.Context) {
	feeds := []string{
		"https://newsng.ng/feed/",
		"https://www.premiumtimesng.com/security/feed",
		"https://www.vanguardngr.com/category/security/feed",
		"https://dailypost.ng/category/security/feed",
		"https://punchng.com/topics/security/feed",
		"https://www.channelstv.com/category/news/security/feed",
	}

	fp := gofeed.NewParser()
	var announcements []models.Announcement

	for _, feedURL := range feeds {
		feed, err := fp.ParseURL(feedURL)
		if err != nil {
			continue
		}

		for _, item := range feed.Items {
			title := strings.ToLower(item.Title)
			desc := strings.ToLower(item.Description)

			securityKeywords := []string{"kidnap", "bandit", "terror", "attack", "security", "alert", "warning", "shooting", "bomb", "explosion", "crisis", "violence", "abduction"}
			isSecurity := false
			for _, kw := range securityKeywords {
				if strings.Contains(title, kw) || strings.Contains(desc, kw) {
					isSecurity = true
					break
				}
			}
			if !isSecurity {
				continue
			}

			// Check if already exists (by title)
			var existing models.Announcement
			if err := DB.Where("title = ?", item.Title).First(&existing).Error; err == nil {
				continue
			}

			severity := "medium"
			if strings.Contains(title, "kidnap") || strings.Contains(title, "terror") || strings.Contains(title, "attack") {
				severity = "high"
			}
			if strings.Contains(title, "critical") || strings.Contains(title, "emergency") {
				severity = "critical"
			}

			var admin models.User
			DB.Where("role = ?", "unit_admin").First(&admin)

			announcement := models.Announcement{
				CreatedBy: admin.ID,
				Title:     item.Title,
				Content:   item.Description,
				Type:      "alert",
				Severity:  severity,
				IsPublic:  true,
			}
			DB.Create(&announcement)
			announcements = append(announcements, announcement)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "RSS monitoring complete",
		"announcements": announcements,
		"count":         len(announcements),
	})
}

// SocialMediaMonitor simulates social media monitoring (mock)
func SocialMediaMonitor(c *gin.Context) {
	mockData := []struct {
		Source  string
		Content string
	}{
		{"Twitter", "Breaking: Kidnapping reported along Abuja-Kaduna highway. Travelers advised to be cautious."},
		{"Facebook", "Community alert: Suspicious persons seen in Ogun State. Report any unusual activity."},
		{"Twitter", "Security update: Bandits attack village in Zamfara, 10 abducted."},
	}

	var announcements []models.Announcement
	var admin models.User
	DB.Where("role = ?", "unit_admin").First(&admin)

	for _, post := range mockData {
		// Check if already exists (simplified)
		var existing models.Announcement
		if err := DB.Where("title LIKE ?", "%"+post.Content[:30]+"%").First(&existing).Error; err == nil {
			continue
		}

		severity := "medium"
		if strings.Contains(strings.ToLower(post.Content), "kidnap") || strings.Contains(strings.ToLower(post.Content), "attack") {
			severity = "high"
		}
		if strings.Contains(strings.ToLower(post.Content), "critical") || strings.Contains(strings.ToLower(post.Content), "emergency") {
			severity = "critical"
		}

		announcement := models.Announcement{
			CreatedBy: admin.ID,
			Title:     "Social Media Alert: " + post.Source,
			Content:   post.Content,
			Type:      "alert",
			Severity:  severity,
			IsPublic:  true,
		}
		DB.Create(&announcement)
		announcements = append(announcements, announcement)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Social media monitoring complete",
		"announcements": announcements,
		"count":         len(announcements),
	})
}

// GetAIAnalysis retrieves AI analysis for a specific source
func GetAIAnalysis(c *gin.Context) {
	sourceType := c.Query("sourceType")
	sourceID := c.Query("sourceId")

	query := DB.Model(&models.AIAnalysis{})
	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}
	if sourceID != "" {
		query = query.Where("source_id = ?", sourceID)
	}

	var analyses []models.AIAnalysis
	if err := query.Order("created_at desc").Find(&analyses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analyses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analyses": analyses})
}