package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

func CreateAnnouncement(c *gin.Context) {
	var input struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		UnitID  string `json:"unitId"`
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

	announcement := models.Announcement{
		Title:     input.Title,
		Content:   input.Content,
		CreatedBy: userObj.ID,
		Status:    "active",
	}

	if input.UnitID != "" {
		unitID, err := uuid.Parse(input.UnitID)
		if err == nil {
			announcement.UnitID = &unitID
		}
	}

	if err := config.DB.Create(&announcement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create announcement"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Announcement created successfully",
		"announcement": announcement,
	})
}

func GetAllAnnouncements(c *gin.Context) {
	var announcements []models.Announcement
	if err := config.DB.Preload("Author").Order("created_at desc").Find(&announcements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch announcements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"announcements": announcements,
	})
}

func GetAnnouncementByID(c *gin.Context) {
	id := c.Param("id")
	announcementID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid announcement ID"})
		return
	}

	var announcement models.Announcement
	if err := config.DB.Preload("Author").First(&announcement, "id = ?", announcementID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"announcement": announcement,
	})
}

func UpdateAnnouncement(c *gin.Context) {
	id := c.Param("id")
	announcementID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid announcement ID"})
		return
	}

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Status  string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var announcement models.Announcement
	if err := config.DB.First(&announcement, "id = ?", announcementID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}

	if input.Title != "" {
		announcement.Title = input.Title
	}
	if input.Content != "" {
		announcement.Content = input.Content
	}
	if input.Status != "" {
		announcement.Status = input.Status
	}

	if err := config.DB.Save(&announcement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update announcement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Announcement updated successfully",
		"announcement": announcement,
	})
}

func DeleteAnnouncement(c *gin.Context) {
	id := c.Param("id")
	announcementID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid announcement ID"})
		return
	}

	if err := config.DB.Delete(&models.Announcement{}, "id = ?", announcementID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete announcement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Announcement deleted successfully",
	})
}
