package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"security-solution/config"
	"security-solution/models"
)

type bufferedWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bufferedWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// IdempotencyMiddleware replays a cached response when the same
// Idempotency-Key header is seen twice for the same user on the same route.
// Missing header = pass-through (no idempotency guarantee).
func IdempotencyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}
		if len(key) > 128 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key too long"})
			c.Abort()
			return
		}

		userValue, exists := c.Get("user")
		if !exists {
			c.Next()
			return
		}
		userObj, ok := userValue.(*models.User)
		if !ok {
			c.Next()
			return
		}

		// Look up existing record
		var existing models.IdempotencyRecord
		err := config.DB.Where("key = ? AND user_id = ?", key, userObj.ID).First(&existing).Error

		if err == nil {
			// Found a record
			if existing.Status == "in_progress" {
				c.JSON(http.StatusConflict, gin.H{
					"error": "a request with this Idempotency-Key is still processing",
				})
				c.Abort()
				return
			}
			// Replay the cached response
			c.Header("X-Idempotent-Replay", "true")
			c.Data(existing.StatusCode, "application/json", []byte(existing.ResponseBody))
			c.Abort()
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.Next()
			return
		}

		// Insert in-progress record. Unique index protects against concurrent dupes.
		record := models.IdempotencyRecord{
			Key:       key,
			UserID:    userObj.ID,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Status:    "in_progress",
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		}
		if err := config.DB.Create(&record).Error; err != nil {
			// Unique violation means another request won the race
			c.JSON(http.StatusConflict, gin.H{
				"error": "a request with this Idempotency-Key is still processing",
			})
			c.Abort()
			return
		}

		// Buffer the response so we can persist it
		bw := &bufferedWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = bw

		c.Next()

		// Only persist 2xx responses as replayable. Errors are not cached.
		status := bw.Status()
		if status >= 200 && status < 300 {
			record.Status = "completed"
			record.StatusCode = status
			record.ResponseBody = bw.body.String()
			config.DB.Save(&record)
		} else {
			// Delete the record so the client can retry with the same key
			config.DB.Unscoped().Delete(&record)
		}
	}
}

// CleanupExpiredIdempotencyRecords is called by the scheduler.
func CleanupExpiredIdempotencyRecords() error {
	return config.DB.
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&models.IdempotencyRecord{}).Error
}

