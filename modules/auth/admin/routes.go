package admin

import (
	"fxserver/pkg/router"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Routes struct {
	handler    *Handler
	middleware *Middleware
}

type RoutesParam struct {
	fx.In
	Handler    *Handler
	Middleware *Middleware
}

func NewRoutes(p RoutesParam) router.RouteRegistrar {
	return &Routes{
		handler:    p.Handler,
		middleware: p.Middleware,
	}
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
	auth := e.Group("/auth/admin")

	// Public routes (no auth required)
	auth.GET("/sso/auth-url", r.handler.GetKeycloakAuthURL)
	auth.POST("/sso/callback", r.handler.HandleKeycloakCallback)
	
	// Protected routes (requires admin token)
	auth.GET("/me", r.handler.GetAdminInfo, r.middleware.VerifyAdminToken())
}
