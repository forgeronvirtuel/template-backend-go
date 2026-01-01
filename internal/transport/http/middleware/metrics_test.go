package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"template-backend-go/internal/observability/metrics"
)

// mockRecorder is a test double that records metric calls for verification.
type mockRecorder struct {
	counters   []counterCall
	histograms []histogramCall
	gauges     []gaugeCall
}

type counterCall struct {
	name   string
	labels metrics.Labels
	delta  int64
}

type histogramCall struct {
	name   string
	labels metrics.Labels
	value  float64
}

type gaugeCall struct {
	name   string
	labels metrics.Labels
	value  float64
	isAdd  bool
}

func (m *mockRecorder) IncCounter(name string, labels metrics.Labels, delta int64) {
	m.counters = append(m.counters, counterCall{name, labels, delta})
}

func (m *mockRecorder) ObserveHistogram(name string, labels metrics.Labels, value float64) {
	m.histograms = append(m.histograms, histogramCall{name, labels, value})
}

func (m *mockRecorder) SetGauge(name string, labels metrics.Labels, value float64) {
	m.gauges = append(m.gauges, gaugeCall{name, labels, value, false})
}

func (m *mockRecorder) AddGauge(name string, labels metrics.Labels, delta float64) {
	m.gauges = append(m.gauges, gaugeCall{name, labels, delta, true})
}

func TestMetrics_WithNoopRecorder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup router with no-op recorder
	router := gin.New()
	router.Use(Metrics(MetricsConfig{
		Recorder:       metrics.NewNoop(),
		MetricsEnabled: true,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not panic and should return 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMetrics_DisabledConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup router with metrics disabled
	router := gin.New()
	router.Use(Metrics(MetricsConfig{
		MetricsEnabled: false,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should work normally
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMetrics_RecordsCorrectMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock := &mockRecorder{}

	// Setup router with mock recorder
	router := gin.New()
	router.Use(Metrics(MetricsConfig{
		Recorder:       mock,
		MetricsEnabled: true,
	}))

	router.GET("/users/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.Param("id")})
	})

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify in-flight gauge was incremented and decremented
	if len(mock.gauges) != 2 {
		t.Fatalf("expected 2 gauge calls (increment and decrement), got %d", len(mock.gauges))
	}

	if mock.gauges[0].name != MetricHTTPRequestsInFlight || mock.gauges[0].value != 1 {
		t.Errorf("expected in-flight +1, got name=%s value=%f", mock.gauges[0].name, mock.gauges[0].value)
	}

	if mock.gauges[1].name != MetricHTTPRequestsInFlight || mock.gauges[1].value != -1 {
		t.Errorf("expected in-flight -1, got name=%s value=%f", mock.gauges[1].name, mock.gauges[1].value)
	}

	// Verify counter was recorded
	if len(mock.counters) != 1 {
		t.Fatalf("expected 1 counter call, got %d", len(mock.counters))
	}

	counter := mock.counters[0]
	if counter.name != MetricHTTPRequestsTotal {
		t.Errorf("expected counter name %s, got %s", MetricHTTPRequestsTotal, counter.name)
	}

	if counter.delta != 1 {
		t.Errorf("expected counter delta 1, got %d", counter.delta)
	}

	// Helper to find label value by key
	findLabel := func(labels metrics.Labels, key string) string {
		for _, l := range labels {
			if l.Key == key {
				return l.Value
			}
		}
		return ""
	}

	// Verify labels use route template, not raw path
	if method := findLabel(counter.labels, "method"); method != "GET" {
		t.Errorf("expected method=GET, got %s", method)
	}

	if route := findLabel(counter.labels, "route"); route != "/users/:id" {
		t.Errorf("expected route=/users/:id (template), got %s", route)
	}

	if status := findLabel(counter.labels, "status"); status != "200" {
		t.Errorf("expected status=200, got %s", status)
	}

	// Verify histogram was recorded
	if len(mock.histograms) != 1 {
		t.Fatalf("expected 1 histogram call, got %d", len(mock.histograms))
	}

	histogram := mock.histograms[0]
	if histogram.name != MetricHTTPRequestDuration {
		t.Errorf("expected histogram name %s, got %s", MetricHTTPRequestDuration, histogram.name)
	}

	// Duration should be in seconds and > 0
	if histogram.value <= 0 {
		t.Errorf("expected positive duration, got %f", histogram.value)
	}

	// Verify histogram labels match counter labels
	if findLabel(histogram.labels, "method") != findLabel(counter.labels, "method") {
		t.Errorf("histogram and counter method labels mismatch")
	}

	if findLabel(histogram.labels, "route") != findLabel(counter.labels, "route") {
		t.Errorf("histogram and counter route labels mismatch")
	}

	if findLabel(histogram.labels, "status") != findLabel(counter.labels, "status") {
		t.Errorf("histogram and counter status labels mismatch")
	}
}

func TestMetrics_UnknownRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock := &mockRecorder{}

	// Setup router with mock recorder
	router := gin.New()
	router.Use(Metrics(MetricsConfig{
		Recorder:       mock,
		MetricsEnabled: true,
	}))

	router.GET("/known", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request to unknown route
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	// Verify counter was recorded with "unknown" route
	if len(mock.counters) != 1 {
		t.Fatalf("expected 1 counter call, got %d", len(mock.counters))
	}

	// Helper to find label value by key
	findLabel := func(labels metrics.Labels, key string) string {
		for _, l := range labels {
			if l.Key == key {
				return l.Value
			}
		}
		return ""
	}

	counter := mock.counters[0]
	if route := findLabel(counter.labels, "route"); route != "unknown" {
		t.Errorf("expected route=unknown for unmatched route, got %s", route)
	}

	if status := findLabel(counter.labels, "status"); status != "404" {
		t.Errorf("expected status=404, got %s", status)
	}
}

func TestMetrics_DifferentStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name           string
		statusCode     int
		expectedStatus string
	}{
		{"Success", http.StatusOK, "200"},
		{"Created", http.StatusCreated, "201"},
		{"BadRequest", http.StatusBadRequest, "400"},
		{"NotFound", http.StatusNotFound, "404"},
		{"InternalServerError", http.StatusInternalServerError, "500"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockRecorder{}

			router := gin.New()
			router.Use(Metrics(MetricsConfig{
				Recorder:       mock,
				MetricsEnabled: true,
			}))

			router.GET("/test", func(c *gin.Context) {
				c.JSON(tc.statusCode, gin.H{"status": "test"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.statusCode {
				t.Errorf("expected status %d, got %d", tc.statusCode, w.Code)
			}

			if len(mock.counters) != 1 {
				t.Fatalf("expected 1 counter call, got %d", len(mock.counters))
			}

			// Helper to find label value by key
			findLabel := func(labels metrics.Labels, key string) string {
				for _, l := range labels {
					if l.Key == key {
						return l.Value
					}
				}
				return ""
			}

			if status := findLabel(mock.counters[0].labels, "status"); status != tc.expectedStatus {
				t.Errorf("expected status=%s, got %s", tc.expectedStatus, status)
			}
		})
	}
}

func TestMetrics_NilRecorderDefaultsToNoop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup router with nil recorder (should default to no-op)
	router := gin.New()
	router.Use(Metrics(MetricsConfig{
		Recorder:       nil, // nil should default to no-op
		MetricsEnabled: true,
	}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request - should not panic
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
