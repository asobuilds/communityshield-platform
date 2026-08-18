package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// AuditMiddleware logs all requests
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Log after request is processed
		userID := uuid.Nil
		userInterface, exists := c.Get("user")
		if exists {
			if user, ok := userInterface.(*models.User); ok {
				userID = user.ID
			}
		}

		// Determine unit ID
		var unitID *uuid.UUID
		if userInterface, exists := c.Get("user"); exists {
			if user, ok := userInterface.(*models.User); ok {
				if user.UnitID != nil {
					unitID = user.UnitID
				}
			}
		}

		// Create audit log
		auditLog := models.AuditLog{
			UserID:     &userID,
			UnitID:     unitID,
			Action:     c.Request.Method,
			Resource:   c.Request.URL.Path,
			Details:    c.Request.URL.RawQuery,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			Status:     "success",
			Severity:   "info",
		}

		if len(c.Errors) > 0 {
			auditLog.Status = "failure"
			auditLog.Severity = "error"
			auditLog.Details = c.Errors.String()
		}

		// Calculate response time
		responseTime := time.Since(startTime)
		if responseTime > 5*time.Second {
			auditLog.Severity = "warning"
		}

		// Only log if user is authenticated or for critical operations
		if userID != uuid.Nil || c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
			config.DB.Create(&auditLog)
		}
	}
}