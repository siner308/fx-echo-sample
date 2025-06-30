package admin

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
	auth := e.Group("/auth/admin")

	// Keycloak SSO routes
	auth.GET("/sso/auth-url", r.handler.GetKeycloakAuthURL)
	auth.POST("/sso/callback", r.handler.HandleKeycloakCallback)
}
