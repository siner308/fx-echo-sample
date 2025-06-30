package user

import (
	"net/http"
	"strings"

	"fxserver/pkg/dto"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// UserContextKey provides type-safe context keys
type UserContextKey string

const (
	UserIDKey    UserContextKey = "user_id"
	UserEmailKey UserContextKey = "user_email"
	UserRoleKey  UserContextKey = "user_role"
	TokenTypeKey UserContextKey = "token_type"
)

type MiddlewareParam struct {
	fx.In
	UserAuthService Service
	Logger          *zap.Logger
}

type Middleware struct {
	userAuthService Service
	logger          *zap.Logger
}

func NewMiddleware(p MiddlewareParam) *Middleware {
	return &Middleware{
		userAuthService: p.UserAuthService,
		logger:          p.Logger,
	}
}

// VerifyAccessToken validates user access tokens
func (m *Middleware) VerifyAccessToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			token := extractTokenFromHeader(echoCtx)
			if token == "" {
				return echoCtx.JSON(http.StatusUnauthorized, dto.NewAuthError("Missing or invalid authorization header"))
			}

			claims, err := m.userAuthService.ValidateAccessToken(token)
			if err != nil {
				m.logger.Warn("Invalid access token", zap.Error(err))
				return echoCtx.JSON(http.StatusUnauthorized, dto.NewAuthError("Invalid or expired token"))
			}

			// Set user information in context using type-safe keys
			echoCtx.Set(string(UserIDKey), claims.UserID)
			echoCtx.Set(string(UserEmailKey), claims.Email)
			echoCtx.Set(string(UserRoleKey), claims.Role)
			echoCtx.Set(string(TokenTypeKey), claims.ID)

			return next(echoCtx)
		}
	}
}

// VerifyAccessTokenOptional validates tokens if present but doesn't require them
func (m *Middleware) VerifyAccessTokenOptional() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			token := extractTokenFromHeader(echoCtx)
			if token != "" {
				claims, err := m.userAuthService.ValidateAccessToken(token)
				if err == nil {
					// Set user information in context if token is valid
					echoCtx.Set(string(UserIDKey), claims.UserID)
					echoCtx.Set(string(UserEmailKey), claims.Email)
					echoCtx.Set(string(UserRoleKey), claims.Role)
					echoCtx.Set(string(TokenTypeKey), claims.ID)
				}
			}

			return next(echoCtx)
		}
	}
}

func extractTokenFromHeader(echoCtx echo.Context) string {
	authHeader := echoCtx.Request().Header.Get("Authorization")
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

// Type-safe helper functions to get user info from context
func GetUserID(echoCtx echo.Context) (int, bool) {
	userID, ok := echoCtx.Get(string(UserIDKey)).(int)
	return userID, ok
}

func GetUserEmail(echoCtx echo.Context) (string, bool) {
	email, ok := echoCtx.Get(string(UserEmailKey)).(string)
	return email, ok
}

func GetUserRole(echoCtx echo.Context) (string, bool) {
	role, ok := echoCtx.Get(string(UserRoleKey)).(string)
	return role, ok
}

func GetTokenType(echoCtx echo.Context) (string, bool) {
	tokenType, ok := echoCtx.Get(string(TokenTypeKey)).(string)
	return tokenType, ok
}