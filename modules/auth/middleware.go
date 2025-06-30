package auth

import (
	"net/http"
	"strings"

	"fxserver/pkg/dto"
	"fxserver/pkg/jwt"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Middleware struct {
	authService Service
	logger      *zap.Logger
}

func NewMiddleware(authService Service, logger *zap.Logger) *Middleware {
	return &Middleware{
		authService: authService,
		logger:      logger,
	}
}

// JWTMiddleware validates access tokens
func (m *Middleware) JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractTokenFromHeader(c)
			if token == "" {
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
					Error: "Missing or invalid authorization header",
				})
			}

			claims, err := m.authService.ValidateAccessToken(token)
			if err != nil {
				m.logger.Warn("Invalid access token", zap.Error(err))
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
					Error: "Invalid or expired token",
				})
			}

			// Set user information in context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.Role)
			c.Set("token_type", claims.ID)

			return next(c)
		}
	}
}

// AdminMiddleware validates admin tokens
func (m *Middleware) AdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractTokenFromHeader(c)
			if token == "" {
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
					Error: "Missing or invalid authorization header",
				})
			}

			claims, err := m.authService.ValidateAdminToken(token)
			if err != nil {
				m.logger.Warn("Invalid admin token", zap.Error(err))
				return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
					Error: "Invalid or expired admin token",
				})
			}

			// Verify admin role
			if claims.Role != "admin" {
				m.logger.Warn("Non-admin user attempted admin access", 
					zap.Int("user_id", claims.UserID),
					zap.String("role", claims.Role))
				return c.JSON(http.StatusForbidden, dto.ErrorResponse{
					Error: "Admin access required",
				})
			}

			// Set user information in context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.Role)
			c.Set("token_type", claims.ID)

			return next(c)
		}
	}
}

// OptionalJWTMiddleware validates tokens if present but doesn't require them
func (m *Middleware) OptionalJWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractTokenFromHeader(c)
			if token != "" {
				claims, err := m.authService.ValidateAccessToken(token)
				if err == nil {
					// Set user information in context if token is valid
					c.Set("user_id", claims.UserID)
					c.Set("user_email", claims.Email)
					c.Set("user_role", claims.Role)
					c.Set("token_type", claims.ID)
				}
			}

			return next(c)
		}
	}
}

func extractTokenFromHeader(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// Helper functions to get user info from context
func GetUserID(c echo.Context) (int, bool) {
	userID, ok := c.Get("user_id").(int)
	return userID, ok
}

func GetUserEmail(c echo.Context) (string, bool) {
	email, ok := c.Get("user_email").(string)
	return email, ok
}

func GetUserRole(c echo.Context) (string, bool) {
	role, ok := c.Get("user_role").(string)
	return role, ok
}