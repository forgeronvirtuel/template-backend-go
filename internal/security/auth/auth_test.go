package auth

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestDisabledAuthenticator(t *testing.T) {
	auth := NewDisabledAuthenticator()
	req := httptest.NewRequest("GET", "/test", nil)

	_, err := auth.Authenticate(context.Background(), req)
	if err != ErrUnauthenticated {
		t.Errorf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestAPIKeyAuthenticator_NoKeys(t *testing.T) {
	auth := NewAPIKeyAuthenticator([]string{})
	req := httptest.NewRequest("GET", "/test", nil)

	_, err := auth.Authenticate(context.Background(), req)
	if err != ErrUnauthenticated {
		t.Errorf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestAPIKeyAuthenticator_BearerValid(t *testing.T) {
	auth := NewAPIKeyAuthenticator([]string{"secret123"})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer secret123")

	identity, err := auth.Authenticate(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if identity.Subject != "api_key" {
		t.Errorf("expected Subject 'api_key', got %q", identity.Subject)
	}
}

func TestAPIKeyAuthenticator_BearerInvalid(t *testing.T) {
	auth := NewAPIKeyAuthenticator([]string{"secret123"})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer wrongkey")

	_, err := auth.Authenticate(context.Background(), req)
	if err != ErrUnauthenticated {
		t.Errorf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestAPIKeyAuthenticator_XAPIKeyValid(t *testing.T) {
	auth := NewAPIKeyAuthenticator([]string{"secret456"})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "secret456")

	identity, err := auth.Authenticate(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if identity.Subject != "api_key" {
		t.Errorf("expected Subject 'api_key', got %q", identity.Subject)
	}
}

func TestAPIKeyAuthenticator_XAPIKeyInvalid(t *testing.T) {
	auth := NewAPIKeyAuthenticator([]string{"secret456"})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "wrongkey")

	_, err := auth.Authenticate(context.Background(), req)
	if err != ErrUnauthenticated {
		t.Errorf("expected ErrUnauthenticated, got %v", err)
	}
}

func TestAPIKeyAuthenticator_MultipleKeys(t *testing.T) {
	auth := NewAPIKeyAuthenticator([]string{"key1", "key2", "key3"})

	tests := []struct {
		name      string
		header    string
		value     string
		wantError bool
	}{
		{"key1 Bearer", "Authorization", "Bearer key1", false},
		{"key2 Bearer", "Authorization", "Bearer key2", false},
		{"key3 X-API-Key", "X-API-Key", "key3", false},
		{"invalid Bearer", "Authorization", "Bearer invalid", true},
		{"no header", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.header != "" {
				req.Header.Set(tt.header, tt.value)
			}

			_, err := auth.Authenticate(context.Background(), req)
			if tt.wantError && err != ErrUnauthenticated {
				t.Errorf("expected ErrUnauthenticated, got %v", err)
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestContextWithIdentity(t *testing.T) {
	ctx := context.Background()
	identity := Identity{Subject: "test_user", Scopes: []string{"read", "write"}}

	ctx = ContextWithIdentity(ctx, identity)
	retrieved, ok := IdentityFromContext(ctx)

	if !ok {
		t.Fatal("expected identity in context")
	}
	if retrieved.Subject != identity.Subject {
		t.Errorf("expected Subject %q, got %q", identity.Subject, retrieved.Subject)
	}
	if len(retrieved.Scopes) != len(identity.Scopes) {
		t.Errorf("expected %d scopes, got %d", len(identity.Scopes), len(retrieved.Scopes))
	}
}

func TestIdentityFromContext_NotPresent(t *testing.T) {
	ctx := context.Background()
	_, ok := IdentityFromContext(ctx)
	if ok {
		t.Error("expected identity not to be in context")
	}
}
