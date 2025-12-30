package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	handler := NewHandler()
	router := NewRouter(handler)

	assert.NotNil(t, router)
}

func TestRouterMiddleware(t *testing.T) {
	handler := NewHandler()
	router := NewRouter(handler)

	// Test that middleware is applied by checking recovery works
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	// Recovery middleware should catch panic and return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRouterRoutes(t *testing.T) {
	handler := NewHandler()
	router := NewRouter(handler)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "root endpoint",
			method:     "GET",
			path:       "/",
			wantStatus: http.StatusOK,
		},
		{
			name:       "health endpoint",
			method:     "GET",
			path:       "/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "hello endpoint",
			method:     "GET",
			path:       "/api/v1/hello",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get user endpoint",
			method:     "GET",
			path:       "/api/v1/users/123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "create user endpoint",
			method:     "POST",
			path:       "/api/v1/users",
			wantStatus: http.StatusBadRequest, // Bad request because no body
		},
		{
			name:       "not found",
			method:     "GET",
			path:       "/nonexistent",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "method not allowed",
			method:     "PUT",
			path:       "/health",
			wantStatus: http.StatusNotFound, // Gin returns 404 for method not allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestRouterAPIv1Group(t *testing.T) {
	handler := NewHandler()
	router := NewRouter(handler)

	// Test all API v1 routes are under /api/v1 prefix
	apiV1Routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/hello"},
		{"GET", "/api/v1/users/123"},
		{"POST", "/api/v1/users"},
	}

	for _, route := range apiV1Routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(route.method, route.path, nil)
			router.ServeHTTP(w, req)

			// Should not return 404 (route exists)
			assert.NotEqual(t, http.StatusNotFound, w.Code,
				"Route %s %s should exist", route.method, route.path)
		})
	}
}

func TestSetupRoutes(t *testing.T) {
	// Test that setupRoutes correctly configures all routes
	router := gin.New()
	handler := NewHandler()
	setupRoutes(router, handler)

	routes := router.Routes()

	// Verify minimum number of routes are registered
	assert.GreaterOrEqual(t, len(routes), 5, "Should have at least 5 routes registered")

	// Check specific routes exist
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+" "+route.Path] = true
	}

	expectedRoutes := []string{
		"GET /",
		"GET /health",
		"GET /api/v1/hello",
		"GET /api/v1/users/:id",
		"POST /api/v1/users",
	}

	for _, expected := range expectedRoutes {
		assert.True(t, routePaths[expected],
			"Route %s should be registered", expected)
	}
}

func TestRouterCORS(t *testing.T) {
	// Test that router handles CORS correctly (if middleware added)
	handler := NewHandler()
	router := NewRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	// Without CORS middleware, should return 404
	// If CORS is added later, this test will need updating
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouterStaticHandlers(t *testing.T) {
	handler := NewHandler()
	router := NewRouter(handler)

	// Test that handlers are properly bound
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func BenchmarkNewRouter(b *testing.B) {
	handler := NewHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRouter(handler)
	}
}

func BenchmarkRouterServeHTTP(b *testing.B) {
	handler := NewHandler()
	router := NewRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
	}
}
