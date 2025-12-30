package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"template-backend-go/internal/config"
	"template-backend-go/internal/server"
	httpTransport "template-backend-go/internal/transport/http"
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
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create HTTP handler and router
	handler := httpTransport.NewHandler()
	router := httpTransport.NewRouter(handler)

	// Create and start server
	srv := server.New(cfg.Server.Address(), router)
	return srv.Start()
}

// GetRunServe returns the runServe function for testing purposes
func GetRunServe() func(*cobra.Command, []string) error {
	return runServe
}
