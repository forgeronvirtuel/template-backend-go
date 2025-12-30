package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	// Channel to capture server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on http://%s", s.httpServer.Addr)
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
		return fmt.Errorf("failed to start server: %w", err)
	case <-quit:
		log.Println("Shutting down server...")
	}

	// Graceful shutdown
	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
		return err
	}

	log.Println("Server stopped")
	return nil
}
