package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"template-backend-go/internal/logging"
)

// Server wraps the HTTP server and manages its lifecycle
type Server struct {
	httpServer *http.Server
}

// New creates a new Server instance
func New(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       120 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
	}
}

// Start starts the HTTP server and handles graceful shutdown
func (s *Server) Start() error {
	logger := logging.Logger()

	// Channel to capture server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		logger.Info("HTTP server starting",
			slog.String("address", s.httpServer.Addr),
		)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-serverErrors:
		logger.Error("Failed to start server", slog.Any("error", err))
		return fmt.Errorf("failed to start server: %w", err)
	case sig := <-quit:
		logger.Info("Shutdown signal received",
			slog.String("signal", sig.String()),
		)
	}

	// Graceful shutdown
	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	logger := logging.Logger()
	logger.Info("Initiating graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Error("Error during server shutdown",
			slog.Any("error", err),
		)
		return err
	}

	logger.Info("Server stopped successfully")
	return nil
}
