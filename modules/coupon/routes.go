package coupon

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

func (r *Routes) RegisterProtectedRoutes(protected *echo.Group) {
	// Coupon routes - each with individual auth middleware
	coupons := protected.Group("/coupons")
	
	// Create coupon - requires auth
	coupons.POST("", r.handler.CreateCoupon, r.authMiddleware.JWTMiddleware())
	
	// Get coupon - requires auth
	coupons.GET("/:id", r.handler.GetCoupon, r.authMiddleware.JWTMiddleware())
	
	// Get coupon by code - requires auth
	coupons.GET("/code/:code", r.handler.GetCouponByCode, r.authMiddleware.JWTMiddleware())
	
	// Update coupon - requires auth
	coupons.PUT("/:id", r.handler.UpdateCoupon, r.authMiddleware.JWTMiddleware())
	
	// Delete coupon - requires auth
	coupons.DELETE("/:id", r.handler.DeleteCoupon, r.authMiddleware.JWTMiddleware())
	
	// List coupons - requires auth
	coupons.GET("", r.handler.ListCoupons, r.authMiddleware.JWTMiddleware())
	
	// Use coupon - requires auth
	coupons.POST("/use", r.handler.UseCoupon, r.authMiddleware.JWTMiddleware())
}