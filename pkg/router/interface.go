package router

import "github.com/labstack/echo/v4"

// RouteRegistrar defines the interface for modules to register their routes
type RouteRegistrar interface {
	RegisterRoutes(e *echo.Echo)
}

// ProtectedRouteRegistrar defines the interface for modules that need authentication
type ProtectedRouteRegistrar interface {
	RegisterProtectedRoutes(protected *echo.Group)
}

// AdminRouteRegistrar defines the interface for modules that need admin routes
type AdminRouteRegistrar interface {
	RegisterAdminRoutes(admin *echo.Group)
}