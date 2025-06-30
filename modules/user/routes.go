package user

import (
	"fxserver/modules/auth"
	"fxserver/pkg/router"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Routes struct {
	handler        *Handler
	authMiddleware *auth.Middleware
}

type RoutesDeps struct {
	fx.In
	Handler        *Handler
	AuthMiddleware *auth.Middleware
}

func NewRoutes(deps RoutesDeps) router.ProtectedRouteRegistrar {
	return &Routes{
		handler:        deps.Handler,
		authMiddleware: deps.AuthMiddleware,
	}
}

func (r *Routes) RegisterProtectedRoutes(api *echo.Group) {
	// User routes - each with individual auth middleware
	users := api.Group("/users")
	
	// Create user - requires auth
	users.POST("", r.handler.CreateUser, r.authMiddleware.JWTMiddleware())
	
	// Get user - requires auth 
	users.GET("/:id", r.handler.GetUser, r.authMiddleware.JWTMiddleware())
	
	// Update user - requires auth
	users.PUT("/:id", r.handler.UpdateUser, r.authMiddleware.JWTMiddleware())
	
	// Delete user - requires auth (could require admin in future)
	users.DELETE("/:id", r.handler.DeleteUser, r.authMiddleware.JWTMiddleware())
	
	// List users - requires auth
	users.GET("", r.handler.ListUsers, r.authMiddleware.JWTMiddleware())
}