package cmd

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"template-backend-go/internal/buildinfo"
	"template-backend-go/internal/config"
	"template-backend-go/internal/health"
	"template-backend-go/internal/logging"
	"template-backend-go/internal/security/auth"
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

	// Configure authentication
	authenticator := configureAuth(logger)

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
		Auth: authenticator,
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

// configureAuth builds an Authenticator based on configuration.
// Implements strict mode: if auth is disabled or misconfigured, returns DisabledAuthenticator (always 401).
func configureAuth(logger *slog.Logger) auth.Authenticator {
	enabled := viper.GetBool("auth.enabled")

	if !enabled {
		logger.Info("Authentication disabled (strict mode: protected routes will return 401)")
		return auth.NewDisabledAuthenticator()
	}

	// Auth is enabled: load API keys
	// Support both single string and list of strings
	var keys []string

	// Try as string slice first
	if viper.IsSet("auth.api_keys") {
		keys = viper.GetStringSlice("auth.api_keys")
	}

	// If empty, try as single string (comma-separated)
	if len(keys) == 0 {
		singleKey := viper.GetString("auth.api_keys")
		if singleKey != "" {
			// Support comma-separated keys in single string
			parts := strings.Split(singleKey, ",")
			for _, p := range parts {
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					keys = append(keys, trimmed)
				}
			}
		}
	}

	if len(keys) == 0 {
		logger.Warn("Authentication enabled but no API keys configured (strict mode: protected routes will return 401)")
		return auth.NewDisabledAuthenticator()
	}

	logger.Info("Authentication enabled",
		slog.Int("key_count", len(keys)),
	)

	return auth.NewAPIKeyAuthenticator(keys)
}
