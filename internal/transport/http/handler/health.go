package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"template-backend-go/internal/health"
)

type HealthHandler struct {
	version   string
	readiness *health.Aggregator
}

type HealthOptions struct {
	Version   string
	Readiness *health.Aggregator
}

func NewHealthHandler(opts HealthOptions) *HealthHandler {
	readiness := opts.Readiness
	if readiness == nil {
		readiness = health.NewReadinessAggregator(nil, 0)
	}

	return &HealthHandler{
		version:   opts.Version,
		readiness: readiness,
	}
}

// Liveness probe: process is up
func (h *HealthHandler) Live(c *gin.Context) {
	resp := gin.H{"status": "ok"}
	if h.version != "" {
		resp["version"] = h.version
	}

	c.JSON(http.StatusOK, resp)
}

// Readiness probe: dependencies are ready
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	summary := h.readiness.Run(ctx)

	statusCode := http.StatusOK
	if !summary.Ready {
		statusCode = http.StatusServiceUnavailable
	}

	resp := gin.H{
		"status": summary.Status,
		"checks": summary.Checks,
	}
	if h.version != "" {
		resp["version"] = h.version
	}

	c.JSON(statusCode, resp)
}
