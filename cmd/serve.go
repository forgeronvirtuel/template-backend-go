package cmd

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"template-backend-go/internal/buildinfo"
	"template-backend-go/internal/config"
	"template-backend-go/internal/health"
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

	// Bind flags to viper
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("server.host", serveCmd.Flags().Lookup("host"))
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

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
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
	srv := server.New(cfg.Server.Address(), h)
	return srv.Start()
}

// GetRunServe returns the runServe function for testing purposes
func GetRunServe() func(*cobra.Command, []string) error {
	return runServe
}
