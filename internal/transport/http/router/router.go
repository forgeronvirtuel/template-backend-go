package router

import (
	"github.com/gin-gonic/gin"

	"template-backend-go/internal/transport/http/handler"
	"template-backend-go/internal/transport/http/middleware"
)

type Deps struct {
	Users  *handler.UsersHandler
	Health *handler.HealthHandler
}

func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())

	// Keep gin.Logger() if you want; ideally replace with structured logger middleware later.
	r.Use(gin.Logger())

	if d.Health == nil {
		panic("router: Health handler is required")
	}

	r.GET("/live", d.Health.Live)
	r.GET("/ready", d.Health.Ready)

	v1 := r.Group("/api/v1")
	{
		if d.Users != nil {
			v1.POST("/users", d.Users.CreateUser)
		}
	}

	return r
}
