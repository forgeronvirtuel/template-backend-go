package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"template-backend-go/internal/observability/metrics"
)

// Metric name constants for HTTP instrumentation.
// These follow Prometheus naming conventions but work with any backend.
const (
	// MetricHTTPRequestsTotal counts total HTTP requests.
	// Type: Counter
	// Labels: method, route, status
	MetricHTTPRequestsTotal = "http_requests_total"

	// MetricHTTPRequestDuration measures request latency distribution.
	// Type: Histogram
	// Labels: method, route, status
	// Unit: seconds
	MetricHTTPRequestDuration = "http_request_duration_seconds"

	// MetricHTTPRequestsInFlight tracks concurrent requests.
	// Type: Gauge
	// Labels: none (or method/route if needed, but typically unlabeled)
	MetricHTTPRequestsInFlight = "http_requests_in_flight"
)

// MetricsConfig holds configuration for the HTTP metrics middleware.
type MetricsConfig struct {
	// Recorder is the metrics backend (Prometheus, OTEL, Datadog, no-op, etc.)
	Recorder metrics.Recorder

	// MetricsEnabled allows disabling metrics collection entirely.
	// When false, the middleware becomes a no-op (zero overhead).
	// When true but Recorder is nil, defaults to metrics.NewNoop().
	MetricsEnabled bool
}

// Metrics returns a Gin middleware that instruments HTTP requests with metrics.
//
// The middleware tracks:
//   - Request counter: labeled by method, route template, status code
//   - Request duration: histogram of latencies in seconds
//   - In-flight requests: gauge of concurrent requests
//
// Label values:
//   - method: HTTP method (GET, POST, PUT, DELETE, etc.)
//   - route: Gin route template (e.g., /users/:id, not /users/123)
//   - status: HTTP status code as string (e.g., "200", "404", "500")
//
// Important characteristics:
//   - Uses c.FullPath() to get route template (low cardinality)
//   - Falls back to "unknown" if route cannot be determined
//   - Safe even if recorder is nil (defaults to no-op)
//   - Does NOT log or store request/response bodies
//   - Does NOT include high cardinality labels (request_id, user_id, IP, query params)
//
// Usage:
//
//	recorder := // ... your metrics backend (Prometheus, OTEL, etc.)
//	router.Use(middleware.Metrics(MetricsConfig{
//	    Recorder: recorder,
//	    MetricsEnabled: true,
//	}))
//
// Or with no-op (metrics disabled):
//
//	router.Use(middleware.Metrics(MetricsConfig{
//	    Recorder: metrics.NewNoop(),
//	    MetricsEnabled: true,
//	}))
//
// Or completely disabled (zero overhead):
//
//	router.Use(middleware.Metrics(MetricsConfig{
//	    MetricsEnabled: false,
//	}))
func Metrics(cfg MetricsConfig) gin.HandlerFunc {
	// If metrics are disabled, return a no-op middleware (zero overhead)
	if !cfg.MetricsEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Default to no-op recorder if nil
	recorder := cfg.Recorder
	if recorder == nil {
		recorder = metrics.NewNoop()
	}

	return func(c *gin.Context) {
		// Track request start time
		start := time.Now()

		// Increment in-flight gauge
		// Note: We use empty labels for in-flight gauge to avoid cardinality explosion.
		// If you need labeled in-flight metrics, add method/route carefully.
		recorder.AddGauge(MetricHTTPRequestsInFlight, nil, 1)

		// Ensure in-flight gauge is decremented on exit
		defer func() {
			recorder.AddGauge(MetricHTTPRequestsInFlight, nil, -1)
		}()

		// Process request
		c.Next()

		// Extract route template (use c.FullPath() for low cardinality)
		// FullPath returns the matched route pattern (e.g., "/users/:id")
		// Falls back to "unknown" if route cannot be determined
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		// Extract method and status
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		// Calculate request duration in seconds
		duration := time.Since(start).Seconds()

		// Create labels for counter and histogram
		// Using slice literal (zero allocation, stack-allocated)
		labels := metrics.Labels{
			{Key: "method", Value: method},
			{Key: "route", Value: route},
			{Key: "status", Value: status},
		}

		// Record metrics
		recorder.IncCounter(MetricHTTPRequestsTotal, labels, 1)
		recorder.ObserveHistogram(MetricHTTPRequestDuration, labels, duration)
	}
}
