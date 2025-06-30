package admin

import (
	"net/http"
	"strings"

	"fxserver/pkg/dto"
	"fxserver/pkg/keycloak"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// AdminContextKey provides type-safe context keys
type AdminContextKey string

const (
	AdminUserKey     AdminContextKey = "admin_user"
	AdminUserIDKey   AdminContextKey = "admin_user_id"
	AdminEmailKey    AdminContextKey = "admin_email"
	AdminRoleKey     AdminContextKey = "admin_role"
	TokenTypeKey     AdminContextKey = "token_type"
	KeycloakUserKey  AdminContextKey = "keycloak_user"
)

type MiddlewareParam struct {
	fx.In
	AdminAuthService Service
	Logger           *zap.Logger
}

type Middleware struct {
	adminAuthService Service
	logger           *zap.Logger
}

func NewMiddleware(p MiddlewareParam) *Middleware {
	return &Middleware{
		adminAuthService: p.AdminAuthService,
		logger:           p.Logger,
	}
}

// VerifyAdminToken validates admin tokens (JWT or Keycloak)
func (m *Middleware) VerifyAdminToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			token := extractTokenFromHeader(echoCtx)
			if token == "" {
				return echoCtx.JSON(http.StatusUnauthorized, dto.NewAuthError("Missing or invalid authorization header"))
			}

			claims, err := m.adminAuthService.ValidateAdminToken(token)
			if err != nil {
				m.logger.Warn("Invalid admin token", zap.Error(err))
				return echoCtx.JSON(http.StatusUnauthorized, dto.NewAuthError("Invalid or expired admin token"))
			}

			// Set user information in context using type-safe keys
			echoCtx.Set(string(AdminUserIDKey), claims.UserID)
			echoCtx.Set(string(AdminEmailKey), claims.Email)
			echoCtx.Set(string(AdminRoleKey), claims.Role)
			echoCtx.Set(string(TokenTypeKey), claims.ID)

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

// Type-safe helper functions to get admin info from context
func GetAdminID(echoCtx echo.Context) (int, bool) {
	userID, ok := echoCtx.Get(string(AdminUserIDKey)).(int)
	return userID, ok
}

func GetAdminEmail(echoCtx echo.Context) (string, bool) {
	email, ok := echoCtx.Get(string(AdminEmailKey)).(string)
	return email, ok
}

func GetAdminRole(echoCtx echo.Context) (string, bool) {
	role, ok := echoCtx.Get(string(AdminRoleKey)).(string)
	return role, ok
}

func GetTokenType(echoCtx echo.Context) (string, bool) {
	tokenType, ok := echoCtx.Get(string(TokenTypeKey)).(string)
	return tokenType, ok
}

func GetKeycloakUser(echoCtx echo.Context) (*keycloak.UserInfo, bool) {
	user, ok := echoCtx.Get(string(KeycloakUserKey)).(*keycloak.UserInfo)
	return user, ok
}
