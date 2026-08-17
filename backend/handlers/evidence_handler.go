package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"security-solution/models"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UploadEvidence(c *gin.Context) {
	caseID := c.Param("id")
	parsedCaseID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}
	var caseItem models.Case
	if err := DB.First(&caseItem, "id = ?", parsedCaseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	description := c.PostForm("description")
	var uploaderID uuid.UUID
	if uid := c.PostForm("uploadedBy"); uid != "" {
		parsed, err := uuid.Parse(uid)
		if err == nil {
			uploaderID = parsed
		}
	}
	if uploaderID == uuid.Nil {
		var user models.User
		if err := DB.First(&user).Error; err == nil {
			uploaderID = user.ID
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Uploader not authenticated"})
			return
		}
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	fileType := "document"
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		fileType = "image"
	case ".mp4", ".avi", ".mov", ".mkv", ".webm":
		fileType = "video"
	case ".mp3", ".wav", ".aac", ".ogg":
		fileType = "audio"
	case ".pdf", ".doc", ".docx", ".txt", ".xls", ".xlsx":
		fileType = "document"
	}
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}
	filename := uuid.New().String() + ext
	filePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	evidence := models.Evidence{
		CaseID:      parsedCaseID,
		FileURL:     "/uploads/" + filename,
		FileType:    fileType,
		Description: description,
		UploadedBy:  uploaderID,
	}
	if err := DB.Create(&evidence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save evidence metadata"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Evidence uploaded successfully",
		"evidence": evidence,
	})
}

func GetEvidence(c *gin.Context) {
	caseID := c.Param("id")
	parsedCaseID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}
	var evidenceList []models.Evidence
	if err := DB.Where("case_id = ?", parsedCaseID).Find(&evidenceList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch evidence"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"evidence": evidenceList})
}