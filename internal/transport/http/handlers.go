package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler contains the HTTP handlers
type Handler struct{}

// NewHandler creates a new HTTP handler
func NewHandler() *Handler {
	return &Handler{}
}

// Health returns the health check handler
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

// Welcome returns the welcome page handler
func (h *Handler) Welcome(c *gin.Context) {
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
}

// Hello returns the hello handler
func (h *Handler) Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
}

// GetUser returns the get user handler
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"user_id": id,
		"name":    "John Doe",
	})
}

// CreateUser returns the create user handler
func (h *Handler) CreateUser(c *gin.Context) {
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
}
