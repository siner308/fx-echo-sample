package coupon

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
	coupons := api.Group("/coupons")

	// Admin-only routes (coupon management)
	coupons.GET("", r.handler.ListCoupons, r.adminMiddleware.VerifyAdminToken())
	coupons.GET("/:id", r.handler.GetCoupon, r.adminMiddleware.VerifyAdminToken())
	coupons.POST("", r.handler.CreateCoupon, r.adminMiddleware.VerifyAdminToken())
	coupons.PUT("/:id", r.handler.UpdateCoupon, r.adminMiddleware.VerifyAdminToken())
	coupons.DELETE("/:id", r.handler.DeleteCoupon, r.adminMiddleware.VerifyAdminToken())

	// User routes (coupon usage)
	coupons.POST("/use", r.handler.UseCoupon, r.userMiddleware.VerifyAccessToken()) // User: use coupon
}
