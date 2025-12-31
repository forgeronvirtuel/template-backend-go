package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"template-backend-go/internal/logging"
)

// StructuredLogger returns a Gin middleware that logs all HTTP requests
// in structured JSON format using slog.
//
// Each request produces a single log entry with:
//   - request_id
//   - method
//   - path
//   - status_code
//   - latency_ms
//   - client_ip
func StructuredLogger() gin.HandlerFunc {
	logger := logging.Logger()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Extract request ID (set by RequestID middleware)
		requestID := ""
		if rid, exists := c.Get(RequestIDKey); exists {
			if ridStr, ok := rid.(string); ok {
				requestID = ridStr
			}
		}

		// Attach request ID to context for downstream use
		ctx := logging.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// Build log attributes
		attrs := []slog.Attr{
			slog.String("request_id", requestID),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status_code", statusCode),
			slog.Int64("latency_ms", latency.Milliseconds()),
			slog.String("client_ip", clientIP),
		}

		// Add error if present
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.String()))
		}

		// Log at INFO level
		logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "HTTP request", attrs...)
	}
}
