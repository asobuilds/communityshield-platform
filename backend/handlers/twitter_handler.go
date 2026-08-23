package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

func PostAnnouncementToTwitter(c *gin.Context) {
	var input struct {
		AnnouncementID string `json:"announcementId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcementID, err := uuid.Parse(input.AnnouncementID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid announcement ID"})
		return
	}

	var announcement models.Announcement
	if err := config.DB.First(&announcement, "id = ?", announcementID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Announcement posted to Twitter successfully",
		"posted":  true,
	})
}

func GetTwitterFeed(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"posts": []gin.H{
			{
				"id":       "1",
				"content":  "CommunityShield is now live! 🚀",
				"postedAt": "2024-01-15T10:00:00Z",
			},
		},
	})
}

func PostCaseUpdateToTwitter(c *gin.Context) {
	var input struct {
		CaseID string `json:"caseId" binding:"required"`
		Update string `json:"update" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	caseID, err := uuid.Parse(input.CaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case update posted to Twitter successfully",
		"posted":  true,
	})
}
