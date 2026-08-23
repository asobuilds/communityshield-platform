package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/models"
)

func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		skipPaths := []string{"/health", "/api/v1/auth/login", "/api/v1/auth/register"}
		for _, path := range skipPaths {
			if c.Request.URL.Path == path {
				return
			}
		}

		userInterface, exists := c.Get("user")
		if !exists {
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			return
		}

		auditLog := models.AuditLog{
			UserID:     user.ID,
			Action:     c.Request.Method,
			EntityType: c.Request.URL.Path,
			EntityID:   c.Request.URL.RawQuery,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			Timestamp:  time.Now(),
		}

		if len(c.Errors) > 0 {
			auditLog.OldValue = c.Errors.String()
		}

		responseTime := time.Since(startTime)
		if responseTime > 5*time.Second {
			auditLog.OldValue = "slow_request"
		}

		config.DB.Create(&auditLog)
	}
}
