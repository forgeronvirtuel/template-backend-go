package http

import (
	"github.com/gin-gonic/gin"
)

// NewRouter creates and configures a new Gin router
func NewRouter(handler *Handler) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup routes
	setupRoutes(router, handler)

	return router
}

// setupRoutes configures all application routes
func setupRoutes(router *gin.Engine, handler *Handler) {
	// Root endpoints
	router.GET("/", handler.Welcome)
	router.GET("/health", handler.Health)

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		v1.GET("/hello", handler.Hello)
		v1.GET("/users/:id", handler.GetUser)
		v1.POST("/users", handler.CreateUser)
	}
}
