package router

import "github.com/labstack/echo/v4"

// RouteRegistrar defines the interface for modules to register their routes
// Each module can choose which middleware to apply per route
type RouteRegistrar interface {
	RegisterRoutes(e *echo.Echo)
}
