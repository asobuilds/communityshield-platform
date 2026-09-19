package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SMSWebhookSignature verifies that incoming SMS webhook calls carry a valid
// HMAC-SHA256 signature of the request body, computed with SMS_WEBHOOK_SECRET.
//
// The provider must send the signature in the header "X-SMS-Signature" as
// lowercase hex. Requests without a valid signature are rejected with 401.
//
// If SMS_WEBHOOK_SECRET is not set, the middleware rejects ALL requests —
// fail-closed, never fail-open.
func SMSWebhookSignature() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := strings.TrimSpace(os.Getenv("SMS_WEBHOOK_SECRET"))
		if secret == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "SMS webhook is not configured"})
			c.Abort()
			return
		}

		provided := strings.ToLower(strings.TrimSpace(c.GetHeader("X-SMS-Signature")))
		if provided == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing signature"})
			c.Abort()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(expected), []byte(provided)) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			c.Abort()
			return
		}

		c.Next()
	}
}
