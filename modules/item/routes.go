package item

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

	// Public item routes (no auth required)
	items := api.Group("/items")
	items.GET("", r.handler.GetItems)           // Get all items (with optional type filter)
	items.GET("/types", r.handler.GetItemTypes) // Get item types info
	items.GET("/:id", r.handler.GetItem)        // Get specific item

	// User inventory routes (user auth required)
	users := api.Group("/users")
	users.GET("/:id/inventory", r.handler.GetUserInventory, r.userMiddleware.VerifyAccessToken()) // Get user inventory

	// Admin item management routes (admin auth required)
	admin := api.Group("/admin")
	adminItems := admin.Group("/items")
	adminItems.POST("", r.handler.CreateItem, r.adminMiddleware.VerifyAdminToken())      // Create item
	adminItems.PUT("/:id", r.handler.UpdateItem, r.adminMiddleware.VerifyAdminToken())  // Update item
	adminItems.DELETE("/:id", r.handler.DeleteItem, r.adminMiddleware.VerifyAdminToken()) // Delete item
}