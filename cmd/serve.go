package cmd

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"template-backend-go/internal/buildinfo"
	"template-backend-go/internal/config"
	"template-backend-go/internal/health"
	"template-backend-go/internal/logging"
	"template-backend-go/internal/server"
	"template-backend-go/internal/transport/http/handler"
	"template-backend-go/internal/transport/http/router"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long:  `Start an HTTP server using Gin framework. The server will listen on the specified host and port.`,
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Define flags for the serve command
	serveCmd.Flags().Uint16P("port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringP("host", "H", "0.0.0.0", "Host to bind to")
	serveCmd.Flags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error)")

	// Bind flags to viper
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("server.host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("log.level", serveCmd.Flags().Lookup("log-level"))
}

func runServe(cmd *cobra.Command, args []string) error {
	// Ensure the provided Cobra command flags are reflected in viper.
	// This matters in tests (and any code) that calls runServe directly.
	if host, err := cmd.Flags().GetString("host"); err == nil {
		viper.Set("server.host", host)
	}
	if port, err := cmd.Flags().GetUint16("port"); err == nil {
		viper.Set("server.port", port)
	}
	if logLevel, err := cmd.Flags().GetString("log-level"); err == nil {
		viper.Set("log.level", logLevel)
	}

	// Initialize structured logging
	logLevel := viper.GetString("log.level")
	if logLevel == "" {
		logLevel = "info"
	}
	logging.Init(logLevel)

	logger := logging.Logger()
	logger.Info("Starting server",
		slog.String("version", buildinfo.Version),
		slog.String("log_level", logLevel),
	)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", slog.Any("error", err))
		return err
	}

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Wire HTTP handlers and router
	readiness := health.NewReadinessAggregator(
		nil,           // add checks here (DB ping, cache ping, etc.)
		1*time.Second, // per-check timeout
	)

	deps := router.Deps{
		Health: handler.NewHealthHandler(handler.HealthOptions{
			Version:   buildinfo.Version,
			Readiness: readiness,
		}),
		// Users is intentionally not wired in this template entrypoint yet.
		// Add your application service and pass handler.NewUsersHandler(...) here.
	}

	h := router.New(deps)

	// Create and start server
	logger.Info("HTTP server configured",
		slog.String("address", cfg.Server.Address()),
	)

	srv := server.New(cfg.Server.Address(), h)
	if err := srv.Start(); err != nil {
		logger.Error("Server stopped with error", slog.Any("error", err))
		return err
	}

	logger.Info("Server shutdown complete")
	return nil
}

// GetRunServe returns the runServe function for testing purposes
func GetRunServe() func(*cobra.Command, []string) error {
	return runServe
}
