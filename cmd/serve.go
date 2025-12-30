package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
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
}

func runServe(cmd *cobra.Command, args []string) error {
	// Get flag values
	port, _ := cmd.Flags().GetUint16("port")
	host, _ := cmd.Flags().GetString("host")

	// Set Gin mode (release, debug, test)
	gin.SetMode(gin.ReleaseMode)

	// Create a new Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup routes
	setupRoutes(router)

	// Server configuration
	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxHeaderBytes:    1 << 20,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Channel to capture server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func(serverErrors chan<- error) {
		log.Printf("Server starting on http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}(serverErrors)

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

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Server stopped")

	return nil
}

func setupRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		v1.GET("/hello", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Hello, World!",
			})
		})

		v1.GET("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"user_id": id,
				"name":    "John Doe",
			})
		})

		v1.POST("/users", func(c *gin.Context) {
			var requestBody map[string]interface{}
			if err := c.ShouldBindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"message": "User created successfully",
				"data":    requestBody,
			})
		})
	}

	// Welcome page
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the API",
			"version": "1.0.0",
			"endpoints": []string{
				"/health",
				"/api/v1/hello",
				"/api/v1/users/:id",
				"/api/v1/users (POST)",
			},
		})
	})
}
