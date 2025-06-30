package payment

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

	// Public payment info routes (no auth required)
	payments := api.Group("/payments")
	payments.GET("/methods", r.handler.GetPaymentMethods)   // Get payment methods
	payments.GET("/statuses", r.handler.GetPaymentStatuses) // Get payment statuses

	// User payment routes (user auth required)
	payments.POST("", r.handler.ProcessPayment, r.userMiddleware.VerifyAccessToken())           // Process payment
	payments.GET("/:id", r.handler.GetPayment, r.userMiddleware.VerifyAccessToken())            // Get payment details

	// User payment history routes
	users := api.Group("/users")
	users.GET("/:id/payments", r.handler.GetUserPayments, r.userMiddleware.VerifyAccessToken())        // Get user payments
	users.GET("/:id/payments/summary", r.handler.GetUserPaymentSummary, r.userMiddleware.VerifyAccessToken()) // Get user payment summary

	// Admin payment management routes (admin auth required)
	admin := api.Group("/admin")
	adminPayments := admin.Group("/payments")
	adminPayments.GET("", r.handler.GetAllPayments, r.adminMiddleware.VerifyAdminToken())                    // Get all payments
	adminPayments.GET("/summary", r.handler.GetPaymentSummary, r.adminMiddleware.VerifyAdminToken())         // Get payment summary
	adminPayments.PUT("/:id/status", r.handler.UpdatePaymentStatus, r.adminMiddleware.VerifyAdminToken())    // Update payment status
	adminPayments.POST("/:id/refund", r.handler.RefundPayment, r.adminMiddleware.VerifyAdminToken())         // Refund payment
}