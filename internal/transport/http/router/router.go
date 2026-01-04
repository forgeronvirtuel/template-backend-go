package router

import (
	"github.com/gin-gonic/gin"

	"template-backend-go/internal/observability/metrics"
	"template-backend-go/internal/security/auth"
	"template-backend-go/internal/transport/http/handler"
	"template-backend-go/internal/transport/http/middleware"
)

type Deps struct {
	Users   *handler.UsersHandler
	Health  *handler.HealthHandler
	Metrics metrics.Recorder   // Optional: metrics backend (nil defaults to no-op)
	Auth    auth.Authenticator // Optional: authenticator (nil defaults to strict disabled mode)
}

func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// Fail fast if required handlers are missing.
	if d.Health == nil {
		panic("router: Health handler is required")
	}

	// Default to no-op metrics if not provided
	metricsRecorder := d.Metrics
	if metricsRecorder == nil {
		metricsRecorder = metrics.NewNoop()
	}

	// Default to strict disabled authenticator if not provided
	// This ensures protected routes return 401 by default (strict mode)
	authenticator := d.Auth
	if authenticator == nil {
		authenticator = auth.NewDisabledAuthenticator()
	}

	// Apply request ID middleware globally
	r.Use(middleware.RequestID())

	// Apply structured logging middleware globally
	r.Use(middleware.StructuredLogger())

	// Apply metrics middleware globally (after request_id, before handlers)
	r.Use(middleware.Metrics(middleware.MetricsConfig{
		Recorder:       metricsRecorder,
		MetricsEnabled: true,
	}))

	// Operational endpoints (public, no auth)
	r.GET("/live", d.Health.Live)
	r.GET("/ready", d.Health.Ready)

	// Business API v1
	v1 := r.Group("/api/v1")
	{
		// Protected routes require authentication
		protected := v1.Group("")
		protected.Use(middleware.Auth(authenticator))
		{
			if d.Users != nil {
				protected.POST("/users", d.Users.CreateUser)
			}
		}
	}

	return r
}
