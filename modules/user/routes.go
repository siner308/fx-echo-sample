package user

import (
	adminauth "fxserver/modules/auth/admin"
	userauth "fxserver/modules/auth/user"
	"fxserver/pkg/router"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Routes struct {
	handler         *Handler
	userMiddleware  *userauth.Middleware
	adminMiddleware *adminauth.Middleware
}

type RoutesParam struct {
	fx.In
	Handler         *Handler
	UserMiddleware  *userauth.Middleware
	AdminMiddleware *adminauth.Middleware
}

func NewRoutes(p RoutesParam) router.RouteRegistrar {
	return &Routes{
		handler:         p.Handler,
		userMiddleware:  p.UserMiddleware,
		adminMiddleware: p.AdminMiddleware,
	}
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api/v1")
	users := api.Group("/users")

	// Public routes (no auth required)
	users.POST("/signup", r.handler.CreateUser) // Public: user signup

	// Admin-only routes (user management)
	users.GET("", r.handler.ListUsers, r.adminMiddleware.VerifyAdminToken()) // Admin only: list all users

	// User routes (self-management)
	users.GET("/:id", r.handler.GetUser, r.userMiddleware.VerifyAccessToken())       // User: get user details (self or public info)
	users.PUT("/:id", r.handler.UpdateUser, r.userMiddleware.VerifyAccessToken())    // User: update own profile
	users.DELETE("/:id", r.handler.DeleteUser, r.userMiddleware.VerifyAccessToken()) // User: delete own account
}
