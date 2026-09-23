package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets conservative security headers on every response.
// Safe for a JSON API + SPA frontend. Do not add CSP that blocks the
// frontend — use report-only defaults.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		h.Set("Permissions-Policy", "geolocation=(self), camera=(), microphone=()")
		// Content-Security-Policy deliberately omitted for an API — no HTML is served.
		c.Next()
	}
}
