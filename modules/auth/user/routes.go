package user

import (
	"fxserver/pkg/router"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Routes struct {
	handler *Handler
}

type RoutesParam struct {
	fx.In
	Handler *Handler
}

func NewRoutes(p RoutesParam) router.RouteRegistrar {
	return &Routes{
		handler: p.Handler,
	}
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
	auth := e.Group("/auth")

	// Public user auth routes
	auth.POST("/login", r.handler.Login)
	auth.POST("/refresh", r.handler.RefreshToken)
}
