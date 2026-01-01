// Package metrics provides a minimal, backend-agnostic metrics abstraction.
//
// This package defines an interface layer that allows plugging any metrics
// backend (Prometheus, OTEL, Datadog, StatsD, etc.) without forcing a vendor
// choice. The template ships with a no-op implementation, ensuring the code
// runs safely even when no metrics system is configured.
//
// Design principles:
//   - Zero external dependencies (stdlib only)
//   - Simple API surface (4 methods)
//   - Safe by default (no-op if not configured)
//   - Low cardinality labels only (no request_id, user_id, raw paths, IPs)
//
// Example usage with hypothetical Prometheus backend:
//
//	import "github.com/prometheus/client_golang/prometheus"
//
//	type prometheusRecorder struct {
//	    counters   map[string]*prometheus.CounterVec
//	    histograms map[string]*prometheus.HistogramVec
//	    gauges     map[string]*prometheus.GaugeVec
//	}
//
//	func (r *prometheusRecorder) IncCounter(name string, labels Labels, delta int64) {
//	    counter := r.counters[name]
//	    counter.With(prometheus.Labels(labels)).Add(float64(delta))
//	}
//
//	func (r *prometheusRecorder) ObserveHistogram(name string, labels Labels, value float64) {
//	    histogram := r.histograms[name]
//	    histogram.With(prometheus.Labels(labels)).Observe(value)
//	}
//
// Example usage with hypothetical OTEL backend:
//
//	import "go.opentelemetry.io/otel/metric"
//
//	type otelRecorder struct {
//	    meter metric.Meter
//	}
//
//	func (r *otelRecorder) IncCounter(name string, labels Labels, delta int64) {
//	    counter, _ := r.meter.Int64Counter(name)
//	    attrs := convertLabelsToAttributes(labels)
//	    counter.Add(context.Background(), delta, metric.WithAttributes(attrs...))
//	}
//
// Example usage with hypothetical Datadog backend:
//
//	import "github.com/DataDog/datadog-go/v5/statsd"
//
//	type datadogRecorder struct {
//	    client *statsd.Client
//	}
//
//	func (r *datadogRecorder) IncCounter(name string, labels Labels, delta int64) {
//	    tags := convertLabelsToTags(labels)
//	    r.client.Count(name, delta, tags, 1)
//	}
package metrics

// Label represents a single key-value pair for metric dimensions.
type Label struct {
	Key   string
	Value string
}

// Labels represents a slice of key-value pairs for metric dimensions.
// Using a slice instead of a map reduces allocations and is more efficient
// for high-throughput scenarios (e.g., HTTP middleware on every request).
//
// Labels should be low cardinality (limited distinct values) to avoid
// metric explosion. Typically used for: method, route, status, environment.
//
// NEVER include high cardinality values such as:
//   - request_id
//   - user_id
//   - raw URL paths with IDs (/users/12345)
//   - query parameters
//   - client IP addresses
//   - timestamps
//
// Performance note: Using []Label instead of map[string]string:
//   - Avoids map allocation on every call
//   - Better for backends like StatsD/Datadog (tags), OTEL (attributes)
//   - Can be stack-allocated for small label counts
//   - Zero allocation with proper capacity pre-sizing
//
// Example usage:
//
//	labels := metrics.Labels{
//	    {Key: "method", Value: "GET"},
//	    {Key: "route", Value: "/users/:id"},
//	    {Key: "status", Value: "200"},
//	}
//
// Or with pre-sized capacity:
//
//	labels := make(metrics.Labels, 0, 3)
//	labels = append(labels, metrics.Label{Key: "method", Value: method})
//	labels = append(labels, metrics.Label{Key: "route", Value: route})
//	labels = append(labels, metrics.Label{Key: "status", Value: status})
type Labels []Label

// Recorder is the interface for recording metrics in a backend-agnostic way.
// Implementations must be safe for concurrent use.
//
// The interface is intentionally minimal to support any metrics backend
// (push or pull based) without exposing backend-specific concepts like
// registries, collectors, or exporters.
type Recorder interface {
	// IncCounter increments a counter metric by delta.
	// Counters are monotonically increasing values (e.g., total requests).
	//
	// Parameters:
	//   name: metric name (e.g., "http_requests_total")
	//   labels: dimensional data (e.g., method="GET", route="/users")
	//   delta: amount to increment (typically 1, but supports batch increments)
	IncCounter(name string, labels Labels, delta int64)

	// ObserveHistogram records a histogram observation (distribution of values).
	// Histograms track distributions of values like latency or size.
	//
	// Parameters:
	//   name: metric name (e.g., "http_request_duration_seconds")
	//   labels: dimensional data
	//   value: observed value in appropriate unit (seconds for duration)
	//
	// For duration histograms, value MUST be in seconds (float64) to follow
	// Prometheus conventions and enable cross-backend compatibility.
	ObserveHistogram(name string, labels Labels, value float64)

	// SetGauge sets a gauge metric to an absolute value.
	// Gauges represent values that can go up or down (e.g., in-flight requests).
	//
	// Parameters:
	//   name: metric name (e.g., "http_requests_in_flight")
	//   labels: dimensional data
	//   value: absolute value to set
	SetGauge(name string, labels Labels, value float64)

	// AddGauge increments or decrements a gauge by delta.
	// Useful for tracking changes (e.g., +1 on request start, -1 on end).
	//
	// Parameters:
	//   name: metric name
	//   labels: dimensional data
	//   delta: amount to add (positive) or subtract (negative)
	AddGauge(name string, labels Labels, delta float64)
}

// noopRecorder is a no-operation implementation that discards all metrics.
// Used when no metrics backend is configured, ensuring negligible overhead.
type noopRecorder struct{}

// NewNoop returns a Recorder that discards all metrics.
// Safe to use in production when metrics are not required or not yet configured.
//
// The no-op recorder has negligible overhead (method calls are inlined by the compiler).
func NewNoop() Recorder {
	return &noopRecorder{}
}

// IncCounter does nothing (no-op).
func (n *noopRecorder) IncCounter(_ string, _ Labels, _ int64) {}

// ObserveHistogram does nothing (no-op).
func (n *noopRecorder) ObserveHistogram(_ string, _ Labels, _ float64) {}

// SetGauge does nothing (no-op).
func (n *noopRecorder) SetGauge(_ string, _ Labels, _ float64) {}

// AddGauge does nothing (no-op).
func (n *noopRecorder) AddGauge(_ string, _ Labels, _ float64) {}
