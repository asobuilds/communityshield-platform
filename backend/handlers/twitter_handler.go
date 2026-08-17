package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"security-solution/models"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
)

// TwitterSearchResult represents a Twitter API v2 search response
type TwitterSearchResult struct {
	Data []struct {
		ID        string `json:"id"`
		Text      string `json:"text"`
		CreatedAt string `json:"created_at"`
		AuthorID  string `json:"author_id"`
	} `json:"data"`
	Includes struct {
		Users []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"users"`
	} `json:"includes"`
	Meta struct {
		NewestID    string `json:"newest_id"`
		OldestID    string `json:"oldest_id"`
		ResultCount int    `json:"result_count"`
	} `json:"meta"`
}

// TwitterUser represents a Twitter user
type TwitterUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

// FetchTwitterAlerts fetches security-related tweets from Twitter API
func FetchTwitterAlerts() ([]models.Announcement, error) {
	bearerToken := os.Getenv("TWITTER_BEARER_TOKEN")
	if bearerToken == "" {
		return nil, fmt.Errorf("TWITTER_BEARER_TOKEN not set")
	}

	// Search query: security keywords in Nigeria
	query := "security OR kidnap OR bandit OR attack OR alert OR warning OR shooting OR bomb OR explosion OR crisis OR violence OR abduction -is:retweet lang:en place_country:NG"
	
	url := fmt.Sprintf("https://api.twitter.com/2/tweets/search/recent?query=%s&max_results=10&tweet.fields=created_at,author_id&user.fields=name,username", query)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Twitter API error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TwitterSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var announcements []models.Announcement

	// Get admin user for created_by
	var admin models.User
	DB.Where("role = ?", "unit_admin").First(&admin)

	for _, tweet := range result.Data {
		// Check if already exists (by title)
		var existing models.Announcement
		if err := DB.Where("title = ?", tweet.Text[:min(len(tweet.Text), 50)]+"...").First(&existing).Error; err == nil {
			continue
		}

		// Determine severity based on keywords
		text := strings.ToLower(tweet.Text)
		severity := "medium"
		if strings.Contains(text, "kidnap") || strings.Contains(text, "terror") || strings.Contains(text, "attack") {
			severity = "high"
		}
		if strings.Contains(text, "critical") || strings.Contains(text, "emergency") || strings.Contains(text, "shooting") {
			severity = "critical"
		}

		// Find author name
		authorName := "Twitter User"
		for _, user := range result.Includes.Users {
			if user.ID == tweet.AuthorID {
				authorName = user.Name + " (@" + user.Username + ")"
				break
			}
		}

		announcement := models.Announcement{
			CreatedBy: admin.ID,
			Title:     fmt.Sprintf("Twitter Alert: %s", tweet.Text[:min(len(tweet.Text), 80)]),
			Content:   tweet.Text + "\n\nSource: Twitter (" + authorName + ")",
			Type:      "alert",
			Severity:  severity,
			IsPublic:  true,
		}
		if err := DB.Create(&announcement).Error; err != nil {
			continue
		}
		announcements = append(announcements, announcement)
	}

	return announcements, nil
}

// MonitorTwitterAlerts is the API handler to trigger Twitter monitoring
func MonitorTwitterAlerts(c *gin.Context) {
	announcements, err := FetchTwitterAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Twitter monitoring complete",
		"count":   len(announcements),
		"alerts":  announcements,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}