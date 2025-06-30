package middleware

import (
	"fxserver/modules/auth"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// MiddlewareChain provides common middleware combinations
type MiddlewareChain struct {
	authMiddleware *auth.Middleware
}

func NewMiddlewareChain(authMiddleware *auth.Middleware) *MiddlewareChain {
	return &MiddlewareChain{
		authMiddleware: authMiddleware,
	}
}

// AuthWithRateLimit combines JWT auth with rate limiting
func (m *MiddlewareChain) AuthWithRateLimit(requestsPerMinute int) []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		m.authMiddleware.JWTMiddleware(),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(requestsPerMinute)),
	}
}

// StrictAuth combines JWT auth with strict rate limiting for sensitive operations
func (m *MiddlewareChain) StrictAuth() []echo.MiddlewareFunc {
	return m.AuthWithRateLimit(5) // 5 requests per minute
}

// RegularAuth combines JWT auth with regular rate limiting
func (m *MiddlewareChain) RegularAuth() []echo.MiddlewareFunc {
	return m.AuthWithRateLimit(30) // 30 requests per minute
}