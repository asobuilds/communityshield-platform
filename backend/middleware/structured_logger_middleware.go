package middleware

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/services"
)

// structuredLogger is a Gin middleware that assigns a request ID to every
// request, echoes it back in the X-Request-Id response header, and emits
// exactly one structured log line per completed request.
//
// Log format is controlled by the LOG_FORMAT env var: "text" (dev) or
// "json" (prod, default). The middleware is fail-open: a logging failure
// (including a panic inside the deferred logger) never blocks the request.
//
// Slow-request detection: SLOW_THRESHOLD_MS (default 500). Requests whose
// duration meets or exceed the threshold are escalated to "warn" level and
// tagged with "slowRequest": true.
//
// Test cases:
//   - Request completes in 100ms, threshold 500 -> level=info, no slowRequest field
//   - Request completes in 700ms, threshold 500 -> level=warn, slowRequest=true
//   - Request 200 in 700ms, threshold 500 -> level=warn (slow escalates from info)
//   - Request 500 in 100ms, threshold 500 -> level=error (status wins over slow)
//   - SLOW_THRESHOLD_MS unset -> default 500 used
//   - SLOW_THRESHOLD_MS=invalid -> default 500 used
//   - Text format: slow line ends with " SLOW"
func structuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- request ID: honor client-supplied X-Request-Id, capped & sanitized ---
		reqID := c.Request.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = uuid.NewString()
		} else {
			reqID = sanitizeRequestID(reqID)
		}
		c.Header("X-Request-Id", reqID)
		c.Set("request_id", reqID)

		start := time.Now()
		c.Next()

		// deferred logging — recover so a marshal/panic can't block the response
		defer func() {
			recover()

			// Skip noisy endpoints. /health and /metrics never log.
			// /api/v1/public/* logs only when status >= 400.
			path := c.Request.URL.Path
			if path == "/health" || path == "/metrics" {
				return
			}
			if len(path) >= 13 && path[:13] == "/api/v1/public" {
				if c.Writer.Status() < 400 {
					return
				}
			}

			durationMs := time.Since(start).Milliseconds()
			status := c.Writer.Status()
			isSlow := durationMs >= slowThresholdMs()
			level := levelForStatus(status, isSlow)
			userID := c.GetString("user_id")
			ip := c.ClientIP()
			userAgent := c.Request.Header.Get("User-Agent")

			// Record in-process metrics for /metrics endpoint.
			services.IncRequest(c.Request.Method, path, status, durationMs)

			format := os.Getenv("LOG_FORMAT")
			if format == "" {
				format = "json"
			}

			var line string
			if format == "text" {
				line = formatText(reqID, level, status, c.Request.Method, path, durationMs, userID, isSlow)
			} else {
				line = formatJSON(reqID, level, status, c.Request.Method, path, durationMs, userID, ip, userAgent, isSlow)
			}
			fmt.Fprintln(os.Stdout, line)
		}()
	}
}

// slowThresholdMs reads SLOW_THRESHOLD_MS from the environment and returns
// the parsed integer. Defaults to 500 on unset or invalid input.
func slowThresholdMs() int64 {
	raw := os.Getenv("SLOW_THRESHOLD_MS")
	if raw == "" {
		return 500
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v < 0 {
		return 500
	}
	return v
}

// sanitizeRequestID caps the client-supplied ID at 64 chars and strips
// anything outside [A-Za-z0-9-_].
func sanitizeRequestID(id string) string {
	re := regexp.MustCompile(`[^A-Za-z0-9-_]`)
	cleaned := re.ReplaceAllString(id, "")
	if len(cleaned) > 64 {
		cleaned = cleaned[:64]
	}
	if cleaned == "" {
		return uuid.NewString()
	}
	return cleaned
}

// levelForStatus maps HTTP status to a log level string. Slow requests
// escalate to "warn" unless the status already demands "error" or "warn".
func levelForStatus(status int, isSlow bool) string {
	if status >= 500 {
		return "error"
	}
	if status >= 400 {
		return "warn"
	}
	if isSlow {
		return "warn"
	}
	return "info"
}

// formatJSON builds the 9-field JSON log line. When isSlow is true the
// "slowRequest" field is emitted as true; otherwise it is omitted.
func formatJSON(reqID, level string, status int, method, path string, durationMs int64, userID, ip, userAgent string, isSlow bool) string {
	rec := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339Nano),
		"requestId":   reqID,
		"userId":      userID,
		"method":      method,
		"path":        path,
		"status":      status,
		"durationMs":  durationMs,
		"ip":          ip,
		"userAgent":   userAgent,
	}
	if isSlow {
		rec["slowRequest"] = true
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return `{"level":"error","msg":"log_marshal_failed"}`
	}
	return string(b)
}

// formatText builds a single-line human-readable log line. When isSlow is
// true the line is suffixed with " SLOW".
func formatText(reqID, level string, status int, method, path string, durationMs int64, userID string, isSlow bool) string {
	ts := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	userField := "-"
	if userID != "" {
		userField = userID
	}
	line := fmt.Sprintf("%s %s [%s] %d %s %dms user=%s", ts, level, reqID, status, method, durationMs, userField)
	if isSlow {
		line += " SLOW"
	}
	return line
}

// StructuredLogger exposes the structured logging middleware.
func StructuredLogger() gin.HandlerFunc {
	return structuredLogger()
}