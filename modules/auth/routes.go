package auth

import (
	"fxserver/pkg/router"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Routes struct {
	handler *Handler
}

func NewRoutes(handler *Handler) router.RouteRegistrar {
	return &Routes{
		handler: handler,
	}
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api/v1")
	
	// Auth routes (public) - with rate limiting for security
	authRoutes := api.Group("/auth")
	
	// Login with rate limiting (more restrictive)
	authRoutes.POST("/login", r.handler.Login, 
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(5)), // 5 requests per minute
	)
	
	// Refresh token with moderate rate limiting
	authRoutes.POST("/refresh", r.handler.RefreshToken,
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10)),
	)
	
	// Admin login with stricter rate limiting
	authRoutes.POST("/admin/login", r.handler.AdminLogin,
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(3)), // 3 requests per minute
	)
}