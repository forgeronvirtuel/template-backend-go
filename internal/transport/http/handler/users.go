package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"template-backend-go/internal/domain/apperr"
	"template-backend-go/internal/transport/http/dto"
	"template-backend-go/internal/transport/http/errmap"
	"template-backend-go/internal/transport/http/middleware"
	"template-backend-go/internal/transport/http/response"
	"template-backend-go/internal/transport/http/validation"
)

type UserService interface {
	CreateUser(email, name string) (UserDTO, error)
}

type UserDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UsersHandler struct {
	users UserService
}

func NewUsersHandler(users UserService) *UsersHandler {
	return &UsersHandler{users: users}
}

func (h *UsersHandler) CreateUser(c *gin.Context) {
	requestID := getRequestID(c)

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		details := validation.FromBindError(err)
		response.Fail(c, http.StatusBadRequest, requestID, string(apperr.CodeValidationError), "Invalid request payload", details)
		return
	}

	user, err := h.users.CreateUser(req.Email, req.Name)
	if err != nil {
		m := errmap.Map(err)
		response.Fail(c, m.HTTPStatus, requestID, m.Code, m.Message, nil)
		return
	}

	response.OK(c, http.StatusCreated, requestID, user)
}

func getRequestID(c *gin.Context) string {
	if v, ok := c.Get(middleware.RequestIDKey); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "unknown"
}
