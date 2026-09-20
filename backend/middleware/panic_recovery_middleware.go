package middleware

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"

	"security-solution/services"
)

// PanicRecovery is a Gin middleware that recovers from panics in any
// handler, reports the panic through the ErrorReporter seam, and returns
// a generic 500 JSON body. The stack trace is never leaked to the client.
//
// It is registered innermost (closest to the handler) so that
// StructuredLogger and AuditMiddleware see the final 500 status rather
// than a stale 200. Gin's built-in Recovery() (active via gin.Default())
// remains the outermost safety net for middleware-level panics.
func PanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Build the error value for reporting.
				var err error
				switch v := r.(type) {
				case error:
					err = v
				default:
					err = &panicError{val: r}
				}

				// Capture stack trace, skipping this function's frames.
				buf := make([]byte, 64*1024)
				n := runtime.Stack(buf, false)
				stack := strings.TrimSpace(string(buf[:n]))

				services.ReportError(c, err, map[string]any{
					"stackTrace": stack,
				})

				log.Printf("PANIC: %v (request_id=%s path=%s)", err, c.GetString("request_id"), c.Request.URL.Path)

				c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}

// panicError wraps a non-error recovered value so it satisfies the
// ErrorReporter interface.
type panicError struct {
	val interface{}
}

func (p *panicError) Error() string {
	return fmt.Sprintf("%v", p.val)
}