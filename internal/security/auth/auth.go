package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// Sentinel errors for authentication/authorization failures.
var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrForbidden       = errors.New("forbidden")
)

// Identity represents an authenticated principal.
type Identity struct {
	Subject string   // Identifier for the authenticated entity (e.g., "api_key", user ID)
	Scopes  []string // Optional: list of granted scopes/permissions
}

// Authenticator validates incoming requests and returns an Identity on success.
// Implementations must be safe for concurrent use.
type Authenticator interface {
	// Authenticate inspects the HTTP request and returns an Identity if valid.
	// Returns ErrUnauthenticated if credentials are missing or invalid.
	// Returns ErrForbidden if credentials are valid but lack required permissions.
	Authenticate(ctx context.Context, r *http.Request) (Identity, error)
}

type contextKey string

const identityKey contextKey = "auth_identity"

// ContextWithIdentity attaches an Identity to the context.
func ContextWithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey, id)
}

// IdentityFromContext extracts the Identity from context.
// Returns (Identity{}, false) if not present.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey).(Identity)
	return id, ok
}

// DisabledAuthenticator always returns ErrUnauthenticated.
// Used when auth is disabled/misconfigured to enforce strict 401 policy.
type DisabledAuthenticator struct{}

// NewDisabledAuthenticator creates an authenticator that always denies access.
func NewDisabledAuthenticator() Authenticator {
	return &DisabledAuthenticator{}
}

// Authenticate always returns ErrUnauthenticated (strict mode).
func (d *DisabledAuthenticator) Authenticate(_ context.Context, _ *http.Request) (Identity, error) {
	return Identity{}, ErrUnauthenticated
}

// APIKeyAuthenticator validates API keys from headers.
// Accepts credentials from:
//   - Authorization: Bearer <key>
//   - X-API-Key: <key>
//
// This is a minimal, stdlib-only implementation suitable for internal services.
// For production, consider more robust solutions (JWT, OAuth2, etc.).
type APIKeyAuthenticator struct {
	allowedKeys map[string]bool
}

// NewAPIKeyAuthenticator creates an authenticator with the given allowed keys.
// If keys is empty, all requests will be rejected (strict mode).
func NewAPIKeyAuthenticator(keys []string) Authenticator {
	allowed := make(map[string]bool, len(keys))
	for _, k := range keys {
		if k != "" {
			allowed[k] = true
		}
	}
	return &APIKeyAuthenticator{allowedKeys: allowed}
}

// Authenticate checks for a valid API key in request headers.
func (a *APIKeyAuthenticator) Authenticate(_ context.Context, r *http.Request) (Identity, error) {
	// Try Authorization: Bearer <key>
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			key := strings.TrimSpace(parts[1])
			if a.allowedKeys[key] {
				return Identity{Subject: "api_key"}, nil
			}
		}
	}

	// Try X-API-Key header
	apiKeyHeader := r.Header.Get("X-API-Key")
	if apiKeyHeader != "" {
		key := strings.TrimSpace(apiKeyHeader)
		if a.allowedKeys[key] {
			return Identity{Subject: "api_key"}, nil
		}
	}

	// No valid credentials found
	return Identity{}, ErrUnauthenticated
}
