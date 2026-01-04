package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"template-backend-go/internal/security/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Meta struct {
		RequestID string `json:"request_id"`
	} `json:"meta"`
}

func TestAuthMiddleware_DisabledAuth(t *testing.T) {
	authenticator := auth.NewDisabledAuthenticator()

	router := gin.New()
	router.Use(RequestID()) // Required for request_id in response
	router.Use(Auth(authenticator))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	var resp errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %s", resp.Error.Code)
	}
}

func TestAuthMiddleware_ValidAPIKey(t *testing.T) {
	authenticator := auth.NewAPIKeyAuthenticator([]string{"valid-key"})

	router := gin.New()
	router.Use(RequestID())
	router.Use(Auth(authenticator))
	router.GET("/protected", func(c *gin.Context) {
		// Verify identity is in context
		identity, ok := auth.IdentityFromContext(c.Request.Context())
		if !ok {
			t.Error("expected identity in context")
		}
		if identity.Subject != "api_key" {
			t.Errorf("expected Subject 'api_key', got %q", identity.Subject)
		}
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidAPIKey(t *testing.T) {
	authenticator := auth.NewAPIKeyAuthenticator([]string{"valid-key"})

	router := gin.New()
	router.Use(RequestID())
	router.Use(Auth(authenticator))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	var resp errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %s", resp.Error.Code)
	}

	if resp.Meta.RequestID == "" {
		t.Error("expected request_id in meta")
	}
}

func TestAuthMiddleware_MissingCredentials(t *testing.T) {
	authenticator := auth.NewAPIKeyAuthenticator([]string{"valid-key"})

	router := gin.New()
	router.Use(RequestID())
	router.Use(Auth(authenticator))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_XAPIKey(t *testing.T) {
	authenticator := auth.NewAPIKeyAuthenticator([]string{"x-api-key-value"})

	router := gin.New()
	router.Use(RequestID())
	router.Use(Auth(authenticator))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("X-API-Key", "x-api-key-value")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_NilAuthenticatorPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when authenticator is nil")
		}
	}()

	Auth(nil)
}
