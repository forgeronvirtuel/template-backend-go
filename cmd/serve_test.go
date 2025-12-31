package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"template-backend-go/internal/health"
	"template-backend-go/internal/transport/http/handler"
	"template-backend-go/internal/transport/http/router"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type testCheck struct {
	name     string
	critical bool
	err      error
}

func (c testCheck) Name() string                  { return c.name }
func (c testCheck) Critical() bool                { return c.critical }
func (c testCheck) Check(_ context.Context) error { return c.err }

func newTestRouter(checks []health.ReadinessCheck) *gin.Engine {
	agg := health.NewReadinessAggregator(checks, 50*time.Millisecond)
	return router.New(router.Deps{
		Health: handler.NewHealthHandler(handler.HealthOptions{
			Version:   "test",
			Readiness: agg,
		}),
	})
}

func TestLive_OK_MinimalJSON(t *testing.T) {
	r := newTestRouter(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, "test", body["version"])
	_, hasChecks := body["checks"]
	assert.False(t, hasChecks)
}

func TestReady_503_WhenNoChecks(t *testing.T) {
	r := newTestRouter(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var body struct {
		Status health.SummaryStatus `json:"status"`
		Checks []health.CheckResult `json:"checks"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, health.SummaryStatusNotReady, body.Status)
	assert.Len(t, body.Checks, 0)
}

func TestReady_200_WhenCriticalCheckPasses(t *testing.T) {
	r := newTestRouter([]health.ReadinessCheck{
		testCheck{name: "db", critical: true, err: nil},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Status health.SummaryStatus `json:"status"`
		Checks []health.CheckResult `json:"checks"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, health.SummaryStatusOK, body.Status)
	require.Len(t, body.Checks, 1)
	assert.Equal(t, "db", body.Checks[0].Name)
	assert.Equal(t, health.CheckStatusOK, body.Checks[0].Status)
	assert.True(t, body.Checks[0].Critical)
}

func TestReady_503_WhenCriticalCheckFails(t *testing.T) {
	r := newTestRouter([]health.ReadinessCheck{
		testCheck{name: "db", critical: true, err: errors.New("down")},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var body struct {
		Status health.SummaryStatus `json:"status"`
		Checks []health.CheckResult `json:"checks"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, health.SummaryStatusNotReady, body.Status)
	require.Len(t, body.Checks, 1)
	assert.Equal(t, "db", body.Checks[0].Name)
	assert.Equal(t, health.CheckStatusFail, body.Checks[0].Status)
	assert.True(t, body.Checks[0].Critical)
	assert.NotEmpty(t, body.Checks[0].Error)
}

func TestReady_200_WhenNonCriticalCheckFails(t *testing.T) {
	r := newTestRouter([]health.ReadinessCheck{
		testCheck{name: "cache", critical: false, err: errors.New("down")},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Status health.SummaryStatus `json:"status"`
		Checks []health.CheckResult `json:"checks"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, health.SummaryStatusOK, body.Status)
	require.Len(t, body.Checks, 1)
	assert.Equal(t, "cache", body.Checks[0].Name)
	assert.Equal(t, health.CheckStatusFail, body.Checks[0].Status)
	assert.False(t, body.Checks[0].Critical)
}
