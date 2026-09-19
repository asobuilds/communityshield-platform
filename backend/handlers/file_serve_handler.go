package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// ServeFile streams a stored file by category and relative hash path.
//
// Access rules:
//   - avatars/* and covers/*  : any authenticated user (public-safe identity)
//   - evidence/*              : requires case access via caseId query param
//   - gov_ids/*               : admin only (super admin / head admin / unit admin)
//   - voice_notes/*           : requires case access via caseId query param
func ServeFile(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	category := c.Param("category")
	hash := c.Param("hash")

	if category == "" || hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category and hash are required"})
		return
	}
	if strings.Contains(hash, "..") || strings.Contains(category, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	switch category {
	case "avatars", "covers":
		// public-safe: any authed user may read
	case "evidence", "voice_notes":
		caseIDStr := c.Query("caseId")
		if caseIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "caseId query param required for evidence"})
			return
		}
		caseID, err := uuid.Parse(caseIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid caseId"})
			return
		}

		accessLevel := c.GetString("case_access_level")
		if accessLevel == "" {
			// fall back to an explicit check
			var membership models.UnitMembership
			var caseObj models.Case
			if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
				return
			}
			isReporter := caseObj.ReportedBy == userObj.ID
			hasMembership := config.DB.
				Where("unit_id = ? AND user_id = ? AND status = ?", caseObj.UnitID, userObj.ID, models.MembershipActive).
				First(&membership).Error == nil
			if !isReporter && !hasMembership && !userObj.IsSuperAdmin && userObj.Role != "super_admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}
	case "gov_ids":
		if !userObj.IsSuperAdmin && userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown category"})
		return
	}

	// Find the full stored path by scanning — the hash is the prefix, extension unknown
	fullRelPath, err := findStoredByHash(category, hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	storageSvc := services.NewFileStorageService()
	reader, err := storageSvc.Open(fullRelPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	defer reader.Close()

	contentType := contentTypeFromExt(fullRelPath)
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "private, max-age=3600")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, reader)
}

// findStoredByHash scans the storage folder for a file whose name starts with the given hash.
func findStoredByHash(category string, hash string) (string, error) {
	storageSvc := services.NewFileStorageService()

	// Try common extensions first
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".pdf", ".mp4", ".webm", ".mov", ".mp3", ".wav", ".ogg", ".m4a", ".bin"} {
		rel := filepath.Join(category, hash+ext)
		if storageSvc.Exists(rel) {
			return rel, nil
		}
	}
	return "", http.ErrMissingFile
}

func contentTypeFromExt(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".m4a":
		return "audio/mp4"
	}
	return "application/octet-stream"
}
