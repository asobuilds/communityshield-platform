package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// privateReadTTL is how long a presigned GET for a private object stays valid.
const privateReadTTL = 10 * time.Minute

// ServeFile resolves a stored file by category and bare hash and hands the
// caller off to the right place.
//
// Stored keys use <category>/<yyyy>/<mm>/<hash>.<ext>. The route only carries
// :category and :hash, so resolution walks the yyyy/mm levels: ListObjectsV2 on
// R2, filepath.Glob on local disk.
//
// Delivery depends on the prefix:
//   - public  (avatars, covers, units, news, community) -> 302 to R2_PUBLIC_BASE
//   - private (evidence, gov_ids, voice_notes)           -> 302 to a short-lived
//     presigned GET URL. A public URL is never returned for these.
//
// Access rules:
//   - avatars/*, covers/*, units/*, news/*, community/*  : any authenticated user
//   - evidence/*, voice_notes/*  : requires case access via caseId query param
//   - gov_ids/*                  : admin only
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

	if !authorizeFileCategory(c, category, userObj) {
		return
	}

	storageSvc := services.NewFileStorageService()
	key, err := storageSvc.FindByHash(category, strings.ToLower(hash))
	if err != nil || key == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	r2 := storageSvc.R2()
	if r2 == nil {
		// Local disk mode: stream straight from disk.
		serveStoredFile(c, storageSvc, key)
		return
	}

	if services.IsPublicStorageKey(key) {
		target := r2.PublicURL(key)
		if target == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "R2_PUBLIC_BASE is not configured"})
			return
		}
		c.Redirect(http.StatusFound, target)
		return
	}

	// Private prefix: short-lived presigned GET instead of a public URL.
	target, err := r2.PresignGet(c.Request.Context(), key, privateReadTTL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	c.Redirect(http.StatusFound, target)
}

// serveStoredFile streams a key out of the local disk backend.
func serveStoredFile(c *gin.Context, storageSvc *services.FileStorageService, key string) {
	reader, err := storageSvc.Open(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	defer reader.Close()

	c.Header("Content-Type", contentTypeFromExt(key))
	c.Header("Cache-Control", "private, max-age=3600")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, reader)
}

// authorizeFileCategory applies the per-category access rules. It writes the
// error response itself and returns false when access is denied.
func authorizeFileCategory(c *gin.Context, category string, userObj *models.User) bool {
	switch category {
	case "avatars", "covers", "units", "news", "community":
		// public-safe: any authed user may read
		return true
	case "evidence", "voice_notes":
		caseIDStr := c.Query("caseId")
		if caseIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "caseId query param required for evidence"})
			return false
		}
		caseID, err := uuid.Parse(caseIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid caseId"})
			return false
		}

		accessLevel := c.GetString("case_access_level")
		if accessLevel != "" {
			return true
		}

		// fall back to an explicit check
		var membership models.UnitMembership
		var caseObj models.Case
		if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
			return false
		}
		isReporter := caseObj.ReportedBy == userObj.ID
		hasMembership := config.DB.
			Where("unit_id = ? AND user_id = ? AND status = ?", caseObj.UnitID, userObj.ID, models.MembershipActive).
			First(&membership).Error == nil
		if !isReporter && !hasMembership && !userObj.IsSuperAdmin && userObj.Role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return false
		}
		return true
	case "gov_ids":
		if !userObj.IsSuperAdmin && userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return false
		}
		return true
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown category"})
		return false
	}
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
