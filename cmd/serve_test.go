package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpTransport "template-backend-go/internal/transport/http"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// setupTestRouter creates a router for testing
func setupTestRouter() *gin.Engine {
	handler := httpTransport.NewHandler()
	return httpTransport.NewRouter(handler)
}

// TestHealthEndpoint tests the /health endpoint
func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.NotNil(t, response["timestamp"])
}

// TestWelcomeEndpoint tests the / endpoint
func TestWelcomeEndpoint(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Welcome to the API", response["message"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.NotNil(t, response["endpoints"])
}

// TestHelloEndpoint tests the /api/v1/hello endpoint
func TestHelloEndpoint(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/hello", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Hello, World!", response["message"])
}

// TestGetUserEndpoint tests the /api/v1/users/:id endpoint
func TestGetUserEndpoint(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name       string
		userID     string
		wantStatus int
		wantUserID string
	}{
		{
			name:       "valid user ID",
			userID:     "123",
			wantStatus: http.StatusOK,
			wantUserID: "123",
		},
		{
			name:       "another valid user ID",
			userID:     "456",
			wantStatus: http.StatusOK,
			wantUserID: "456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/v1/users/"+tt.userID, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.wantUserID, response["user_id"])
			assert.Equal(t, "John Doe", response["name"])
		})
	}
}

// TestCreateUserEndpoint tests the /api/v1/users POST endpoint
func TestCreateUserEndpoint(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name       string
		payload    interface{}
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid user creation",
			payload: map[string]interface{}{
				"name":  "Jane Doe",
				"email": "jane@example.com",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "valid user with additional fields",
			payload: map[string]interface{}{
				"name":  "John Smith",
				"email": "john@example.com",
				"age":   30,
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name:       "invalid JSON",
			payload:    "invalid json",
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "empty body",
			payload:    nil,
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tt.payload == nil {
				body = bytes.NewBuffer(nil)
			} else if str, ok := tt.payload.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				jsonData, _ := json.Marshal(tt.payload)
				body = bytes.NewBuffer(jsonData)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/users", body)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.wantError {
				assert.NotNil(t, response["error"])
			} else {
				assert.Equal(t, "User created successfully", response["message"])
				assert.NotNil(t, response["data"])
			}
		})
	}
}

// TestNotFoundEndpoint tests 404 handling
func TestNotFoundEndpoint(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestMethodNotAllowed tests unsupported HTTP methods
func TestMethodNotAllowed(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "PUT on GET-only endpoint",
			method: "PUT",
			path:   "/health",
		},
		{
			name:   "DELETE on GET-only endpoint",
			method: "DELETE",
			path:   "/api/v1/hello",
		},
		{
			name:   "GET on POST-only endpoint",
			method: "GET",
			path:   "/api/v1/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			// Gin returns 404 for method not allowed by default
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	}
}

// BenchmarkHealthEndpoint benchmarks the /health endpoint
func BenchmarkHealthEndpoint(b *testing.B) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkGetUserEndpoint benchmarks the /api/v1/users/:id endpoint
func BenchmarkGetUserEndpoint(b *testing.B) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkCreateUserEndpoint benchmarks the /api/v1/users POST endpoint
func BenchmarkCreateUserEndpoint(b *testing.B) {
	router := setupTestRouter()

	payload := map[string]interface{}{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}
	jsonData, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}
