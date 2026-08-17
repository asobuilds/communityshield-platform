package handlers

import (
	"net/http"
	"security-solution/models"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateAnnouncement(c *gin.Context) {
	var input struct {
		UnitID    string  `json:"unitId"`
		Title     string  `json:"title" binding:"required"`
		Content   string  `json:"content" binding:"required"`
		Type      string  `json:"type" binding:"required"`
		Severity  string  `json:"severity" binding:"required"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		RadiusKM  int     `json:"radiusKm"`
		IsPublic  bool    `json:"isPublic"`
		ExpiresAt string  `json:"expiresAt"`
		UserID    string  `json:"userId" binding:"required"`
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

	var unitID *uuid.UUID
	if input.UnitID != "" {
		parsed, err := uuid.Parse(input.UnitID)
		if err == nil {
			unitID = &parsed
		}
	}

	var expiresAt *time.Time
	if input.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, input.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	announcement := models.Announcement{
		UnitID:    unitID,
		CreatedBy: userID,
		Title:     input.Title,
		Content:   input.Content,
		Type:      input.Type,
		Severity:  input.Severity,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
		RadiusKM:  input.RadiusKM,
		IsPublic:  input.IsPublic,
		ExpiresAt: expiresAt,
	}

	if err := DB.Create(&announcement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create announcement"})
		return
	}

	// 🔥 Send push notifications to all users if public
	if input.IsPublic {
		// Get all users
		var users []models.User
		if err := DB.Find(&users).Error; err == nil {
			title := "📢 " + input.Title
			body := input.Content
			if len(body) > 100 {
				body = body[:100] + "..."
			}
			for _, u := range users {
				SendPushNotification(u.ID, title, body, "/vite.svg", "/alerts")
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Announcement created",
		"announcement": announcement,
	})
}

func GetAnnouncements(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.Query("radius")
	severity := c.Query("severity")
	typeFilter := c.Query("type")
	publicOnly := c.Query("public")

	query := DB.Model(&models.Announcement{})

	if publicOnly == "true" {
		query = query.Where("is_public = ?", true)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}
	query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())

	var announcements []models.Announcement
	if err := query.Order("created_at desc").Find(&announcements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch announcements"})
		return
	}

	if latStr != "" && lngStr != "" {
		lat, err1 := strconv.ParseFloat(latStr, 64)
		lng, err2 := strconv.ParseFloat(lngStr, 64)
		if err1 == nil && err2 == nil {
			radius := 10.0
			if r, err := strconv.ParseFloat(radiusStr, 64); err == nil && r > 0 {
				radius = r
			}
			_ = radius
			filtered := []models.Announcement{}
			for _, a := range announcements {
				if a.Latitude == nil || a.Longitude == nil {
					if a.RadiusKM == 0 {
						filtered = append(filtered, a)
					}
					continue
				}
				dist := haversine(lat, lng, *a.Latitude, *a.Longitude)
				if dist <= float64(a.RadiusKM) || a.RadiusKM == 0 {
					filtered = append(filtered, a)
				}
			}
			announcements = filtered
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"announcements": announcements,
		"count":         len(announcements),
	})
}

func GetAnnouncement(c *gin.Context) {
	id := c.Param("id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var ann models.Announcement
	if err := DB.First(&ann, "id = ?", parsed).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"announcement": ann})
}

func UpdateAnnouncement(c *gin.Context) {
	id := c.Param("id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var input struct {
		Title     string  `json:"title"`
		Content   string  `json:"content"`
		Type      string  `json:"type"`
		Severity  string  `json:"severity"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		RadiusKM  int     `json:"radiusKm"`
		IsPublic  bool    `json:"isPublic"`
		ExpiresAt string  `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var ann models.Announcement
	if err := DB.First(&ann, "id = ?", parsed).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}
	if input.Title != "" {
		ann.Title = input.Title
	}
	if input.Content != "" {
		ann.Content = input.Content
	}
	if input.Type != "" {
		ann.Type = input.Type
	}
	if input.Severity != "" {
		ann.Severity = input.Severity
	}
	ann.Latitude = input.Latitude
	ann.Longitude = input.Longitude
	if input.RadiusKM > 0 {
		ann.RadiusKM = input.RadiusKM
	}
	ann.IsPublic = input.IsPublic
	if input.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, input.ExpiresAt); err == nil {
			ann.ExpiresAt = &t
		}
	}
	if err := DB.Save(&ann).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":      "Announcement updated",
		"announcement": ann,
	})
}

func DeleteAnnouncement(c *gin.Context) {
	id := c.Param("id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := DB.Delete(&models.Announcement{}, "id = ?", parsed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Announcement deleted"})
}