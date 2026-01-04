package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"template-backend-go/internal/domain/apperr"
	"template-backend-go/internal/logging"
	"template-backend-go/internal/security/auth"
	"template-backend-go/internal/transport/http/response"
)

// Auth returns a Gin middleware that enforces authentication using the provided Authenticator.
//
// On success:
//   - Attaches the Identity to the request context
//   - Allows the request to proceed
//
// On failure:
//   - Returns a business API error response (error/meta format)
//   - Maps auth.ErrUnauthenticated -> 401 UNAUTHORIZED
//   - Maps auth.ErrForbidden -> 403 FORBIDDEN
//   - Other errors -> 500 INTERNAL_ERROR
//   - Aborts the request chain
//
// IMPORTANT:
//   - This middleware must NOT be applied to operational endpoints (/live, /ready)
//   - Does not log credentials/tokens (security policy)
//   - Uses existing response.Fail for consistent error format
//
// Usage:
//
//	authenticator := auth.NewAPIKeyAuthenticator(keys)
//	protected := router.Group("/api/v1")
//	protected.Use(middleware.Auth(authenticator))
func Auth(authenticator auth.Authenticator) gin.HandlerFunc {
	// Fail fast if authenticator is nil (programming error)
	if authenticator == nil {
		panic("auth middleware: authenticator cannot be nil")
	}

	return func(c *gin.Context) {
		logger := logging.LoggerFromContext(c.Request.Context())
		requestID := getRequestID(c)

		// Attempt authentication
		identity, err := authenticator.Authenticate(c.Request.Context(), c.Request)
		if err != nil {
			// Map authentication errors to business API errors
			var appErr *apperr.Error

			if errors.Is(err, auth.ErrUnauthenticated) {
				appErr = apperr.New(apperr.CodeUnauthorized, "Authentication required")
				logger.Warn("Authentication failed",
					slog.String("request_id", requestID),
					slog.String("error_code", string(apperr.CodeUnauthorized)),
				)
			} else if errors.Is(err, auth.ErrForbidden) {
				appErr = apperr.New(apperr.CodeForbidden, "Insufficient permissions")
				logger.Warn("Authorization failed",
					slog.String("request_id", requestID),
					slog.String("error_code", string(apperr.CodeForbidden)),
				)
			} else {
				// Unexpected error during authentication
				appErr = apperr.New(apperr.CodeInternal, "Authentication error")
				logger.Error("Unexpected authentication error",
					slog.String("request_id", requestID),
					slog.Any("error", err),
				)
			}

			// Return business API error response
			response.Fail(c, mapHTTPStatus(appErr.Code), requestID, string(appErr.Code), appErr.Message, nil)
			c.Abort()
			return
		}

		// Authentication successful: attach identity to context
		ctx := auth.ContextWithIdentity(c.Request.Context(), identity)
		c.Request = c.Request.WithContext(ctx)

		// Log successful authentication (without exposing sensitive data)
		logger.Debug("Request authenticated",
			slog.String("request_id", requestID),
			slog.String("subject", identity.Subject),
		)

		c.Next()
	}
}

// mapHTTPStatus maps apperr.Code to HTTP status code.
// Duplicates some logic from errmap package but kept here for clarity.
func mapHTTPStatus(code apperr.Code) int {
	switch code {
	case apperr.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperr.CodeForbidden:
		return http.StatusForbidden
	case apperr.CodeNotFound:
		return http.StatusNotFound
	case apperr.CodeConflict:
		return http.StatusConflict
	case apperr.CodeValidationError:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// getRequestID extracts the request ID from Gin context.
// Returns "unknown" if not found (should not happen if RequestID middleware is used).
func getRequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "unknown"
}
