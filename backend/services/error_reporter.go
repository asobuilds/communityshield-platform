package services

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorReporter is the out-of-band error reporting seam. Concrete
// implementations (Sentry, GlitchTip, self-hosted) are wired from main.go;
// until then the NoopReporter is the default and nothing is shipped.
type ErrorReporter interface {
	Report(err error, fields map[string]any)
}

// noopReporter is the zero-value default. It intentionally does nothing.
type noopReporter struct{}

func (n *noopReporter) Report(err error, fields map[string]any) {}

// SetErrorReporter installs the active reporter. Call once at startup.
func SetErrorReporter(r ErrorReporter) {
	if r == nil {
		r = &noopReporter{}
	}
	errorReporter = r
}

var errorReporter ErrorReporter = &noopReporter{}

// tokenBucket is a simple in-process rate limiter for error reports.
// Refill: 100 tokens/min. Burst capacity: 200. When empty, the caller
// is dropped and the dropped counter is incremented. No circuit breaker.
type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	lastRefill time.Time
	maxTokens  float64
	rate       float64 // tokens per second
	dropped    int64
}

func newTokenBucket() *tokenBucket {
	return &tokenBucket{
		tokens:     200,
		maxTokens:  200,
		rate:       100.0 / 60.0, // 100 per minute
		lastRefill: time.Now(),
	}
}

// allow returns true if a token is available, otherwise increments the
// dropped counter and returns false.
func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	b.dropped++
	return false
}

var reporterBucket = newTokenBucket()

// ReportError is the package-level helper handlers call explicitly on 5xx
// paths (DB failures, upstream timeouts, etc.). It assembles the standard
// field bag and dispatches to the installed reporter. Returns early on
// nil err or nil context — silent no-op.
func ReportError(c *gin.Context, err error, extra map[string]any) {
	if err == nil || c == nil {
		return
	}
	if !reporterBucket.allow() {
		return
	}

	fields := map[string]any{
		"requestId":    c.GetString("request_id"),
		"userId":       c.GetString("user_id"),
		"path":         c.Request.URL.Path,
		"method":       c.Request.Method,
		"status":       c.Writer.Status(),
		"errorMessage": err.Error(),
		"errorType":    errorTypeName(err),
		"stackTrace":   captureStack(2),
		"timestamp":    time.Now().UTC().Format(time.RFC3339Nano),
	}
	for k, v := range extra {
		fields[k] = v
	}

	errorReporter.Report(err, fields)
	log.Printf("ERROR: %v (request_id=%s path=%s)", err, fields["requestId"], fields["path"])
}

// errorTypeName returns the concrete type name of an error, unwrapping
// the chain via errors.Unwrap so the field is meaningful downstream.
func errorTypeName(err error) string {
	if err == nil {
		return ""
	}
	for {
		w, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		if inner := w.Unwrap(); inner != nil {
			err = inner
			continue
		}
		break
	}
	return fmt.Sprintf("%T", err)
}

// captureStack returns a trimmed stack trace string, skipping the frames
// belonging to this package so the trace starts at the caller.
func captureStack(skip int) string {
	buf := make([]byte, 64*1024)
	n := runtime.Stack(buf, false)
	lines := strings.Split(string(buf[:n]), "\n")
	// Stack layout: 2 lines per frame. Skip the first `skip*2` lines
	// (this function + its caller) plus the reporter frames.
	start := skip*2 + 4
	if start >= len(lines) {
		start = 0
	}
	return strings.Join(lines[start:], "\n")
}

// ErrRateLimitDropped is returned to callers when the bucket is empty.
// It is informational; the report was intentionally not sent.
var ErrRateLimitDropped = errors.New("error report dropped: rate limit exceeded")