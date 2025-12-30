package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewHandler(t *testing.T) {
	handler := NewHandler()
	assert.NotNil(t, handler)
}

func TestHandlerHealth(t *testing.T) {
	handler := NewHandler()
	router := gin.New()
	router.GET("/health", handler.Health)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.NotNil(t, response["timestamp"])

	// Verify timestamp is recent (within last 5 seconds)
	timestamp, ok := response["timestamp"].(float64)
	require.True(t, ok, "timestamp should be a number")
	now := time.Now().Unix()
	assert.InDelta(t, float64(now), timestamp, 5.0)
}

func TestHandlerWelcome(t *testing.T) {
	handler := NewHandler()
	router := gin.New()
	router.GET("/", handler.Welcome)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Welcome to the API", response["message"])
	assert.Equal(t, "1.0.0", response["version"])

	endpoints, ok := response["endpoints"].([]interface{})
	require.True(t, ok, "endpoints should be an array")
	assert.Len(t, endpoints, 4)
	assert.Contains(t, endpoints, "/health")
	assert.Contains(t, endpoints, "/api/v1/hello")
}

func TestHandlerHello(t *testing.T) {
	handler := NewHandler()
	router := gin.New()
	router.GET("/hello", handler.Hello)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hello", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Hello, World!", response["message"])
}

func TestHandlerGetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		expectedUserID string
		expectedName   string
	}{
		{
			name:           "numeric user ID",
			userID:         "123",
			expectedUserID: "123",
			expectedName:   "John Doe",
		},
		{
			name:           "string user ID",
			userID:         "abc",
			expectedUserID: "abc",
			expectedName:   "John Doe",
		},
		{
			name:           "UUID user ID",
			userID:         "550e8400-e29b-41d4-a716-446655440000",
			expectedUserID: "550e8400-e29b-41d4-a716-446655440000",
			expectedName:   "John Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler()
			router := gin.New()
			router.GET("/users/:id", handler.GetUser)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedUserID, response["user_id"])
			assert.Equal(t, tt.expectedName, response["name"])
		})
	}
}

func TestHandlerCreateUser(t *testing.T) {
	tests := []struct {
		name       string
		payload    interface{}
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid user data",
			payload: map[string]interface{}{
				"name":  "Jane Doe",
				"email": "jane@example.com",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "valid user with extra fields",
			payload: map[string]interface{}{
				"name":    "John Smith",
				"email":   "john@example.com",
				"age":     30,
				"address": "123 Main St",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name:       "invalid JSON",
			payload:    "invalid json string",
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "empty body",
			payload:    nil,
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "empty object",
			payload:    map[string]interface{}{},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler()
			router := gin.New()
			router.POST("/users", handler.CreateUser)

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
			req := httptest.NewRequest("POST", "/users", body)
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

func TestHandlerCreateUserContentType(t *testing.T) {
	handler := NewHandler()
	router := gin.New()
	router.POST("/users", handler.CreateUser)

	payload := map[string]interface{}{
		"name": "Test User",
	}
	jsonData, _ := json.Marshal(payload)

	tests := []struct {
		name        string
		contentType string
		wantStatus  int
	}{
		{
			name:        "with content-type",
			contentType: "application/json",
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "without content-type",
			contentType: "",
			wantStatus:  http.StatusCreated, // Gin accepts JSON without explicit Content-Type
		},
		{
			name:        "wrong content-type",
			contentType: "text/plain",
			wantStatus:  http.StatusCreated, // Gin still parses valid JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func BenchmarkHandlerHealth(b *testing.B) {
	handler := NewHandler()
	router := gin.New()
	router.GET("/health", handler.Health)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHandlerGetUser(b *testing.B) {
	handler := NewHandler()
	router := gin.New()
	router.GET("/users/:id", handler.GetUser)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users/123", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHandlerCreateUser(b *testing.B) {
	handler := NewHandler()
	router := gin.New()
	router.POST("/users", handler.CreateUser)

	payload := map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
	}
	jsonData, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}
