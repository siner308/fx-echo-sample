package reward

import (
	adminauth "fxserver/modules/auth/admin"
	"fxserver/pkg/router"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Routes struct {
	handler         *Handler
	adminMiddleware *adminauth.Middleware
}

type RoutesParam struct {
	fx.In
	Handler         *Handler
	AdminMiddleware *adminauth.Middleware
}

func NewRoutes(p RoutesParam) router.RouteRegistrar {
	return &Routes{
		handler:         p.Handler,
		adminMiddleware: p.AdminMiddleware,
	}
}

func (r *Routes) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api/v1")

	// Public reward info routes (no auth required)
	rewards := api.Group("/rewards")
	rewards.GET("/sources", r.handler.GetRewardSources) // Get reward sources info

	// Admin reward management routes (admin auth required)
	admin := api.Group("/admin")
	adminRewards := admin.Group("/rewards")
	adminRewards.POST("/grant", r.handler.GrantReward, r.adminMiddleware.VerifyAdminToken())      // Grant reward to single user
	adminRewards.POST("/bulk-grant", r.handler.BulkGrantReward, r.adminMiddleware.VerifyAdminToken()) // Grant rewards to multiple users
}