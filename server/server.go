package server

import (
	"context"
	"net/http"

	"fxserver/middleware"
	"fxserver/modules/coupon"
	"fxserver/modules/user"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type EchoServer struct {
	echo *echo.Echo
	log  *zap.Logger
}

func NewEchoServer(
	lc fx.Lifecycle,
	log *zap.Logger,
	loggerMiddleware *middleware.LoggerMiddleware,
	errorMiddleware *middleware.ErrorMiddleware,
	userHandler *user.Handler,
	couponHandler *coupon.Handler,
) *EchoServer {
	e := echo.New()

	// Set error handler
	e.HTTPErrorHandler = errorMiddleware.ErrorHandler()

	// Add middleware
	e.Use(loggerMiddleware.LoggerMiddleware())

	// Setup routes
	setupRoutes(e, userHandler, couponHandler)

	server := &EchoServer{
		echo: e,
		log:  log,
	}

	lc.Append(fx.Hook{
		OnStart: server.Start,
		OnStop:  server.Stop,
	})

	return server
}

func (s *EchoServer) Start(ctx context.Context) error {
	s.log.Info("Starting HTTP server", zap.String("addr", ":8080"))
	go func() {
		if err := s.echo.Start(":8080"); err != nil && err != http.ErrServerClosed {
			s.log.Fatal("Server failed to start", zap.Error(err))
		}
	}()
	return nil
}

func (s *EchoServer) Stop(ctx context.Context) error {
	s.log.Info("Stopping HTTP server")
	return s.echo.Shutdown(ctx)
}

func setupRoutes(e *echo.Echo, userHandler *user.Handler, couponHandler *coupon.Handler) {
	api := e.Group("/api/v1")

	// User routes
	users := api.Group("/users")
	users.POST("", userHandler.CreateUser)
	users.GET("/:id", userHandler.GetUser)
	users.PUT("/:id", userHandler.UpdateUser)
	users.DELETE("/:id", userHandler.DeleteUser)
	users.GET("", userHandler.ListUsers)

	// Coupon routes
	coupons := api.Group("/coupons")
	coupons.POST("", couponHandler.CreateCoupon)
	coupons.GET("/:id", couponHandler.GetCoupon)
	coupons.GET("/code/:code", couponHandler.GetCouponByCode)
	coupons.PUT("/:id", couponHandler.UpdateCoupon)
	coupons.DELETE("/:id", couponHandler.DeleteCoupon)
	coupons.GET("", couponHandler.ListCoupons)
	coupons.POST("/use", couponHandler.UseCoupon)
}
