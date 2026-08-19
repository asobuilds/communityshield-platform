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

		// Skip logging for health and public endpoints
		skipPaths := []string{"/health", "/api/v1/auth/login", "/api/v1/auth/register"}
		for _, path := range skipPaths {
			if c.Request.URL.Path == path {
				return
			}
		}

		// Only log authenticated requests
		userInterface, exists := c.Get("user")
		if !exists {
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			return
		}

		// Determine unit ID
		var unitID *uuid.UUID
		if user.UnitID != nil {
			unitID = user.UnitID
		}

		// Create audit log
		auditLog := models.AuditLog{
			UserID:    &user.ID,
			UnitID:    unitID,
			Action:    c.Request.Method,
			Resource:  c.Request.URL.Path,
			Details:   c.Request.URL.RawQuery,
			IPAddress: c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Status:    "success",
			Severity:  "info",
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

		config.DB.Create(&auditLog)
	}
}