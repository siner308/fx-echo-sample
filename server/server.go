package server

import (
	"context"
	"net/http"

	"fxserver/middleware"
	"fxserver/modules/auth"
	"fxserver/pkg/router"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type EchoServer struct {
	echo *echo.Echo
	log  *zap.Logger
}

type ServerDeps struct {
	fx.In
	Lifecycle           fx.Lifecycle
	Logger              *zap.Logger
	LoggerMiddleware    *middleware.LoggerMiddleware
	ErrorMiddleware     *middleware.ErrorMiddleware
	AuthMiddleware      *auth.Middleware
	RouteRegistrars     []router.RouteRegistrar           `group:"routes"`
	ProtectedRegistrars []router.ProtectedRouteRegistrar  `group:"protected_routes"`
	AdminRegistrars     []router.AdminRouteRegistrar      `group:"admin_routes,optional"`
}

func NewEchoServer(deps ServerDeps) *EchoServer {
	e := echo.New()

	// Set error handler
	e.HTTPErrorHandler = deps.ErrorMiddleware.ErrorHandler()

	// Add middleware
	e.Use(deps.LoggerMiddleware.LoggerMiddleware())

	// Setup routes using modular registration
	setupModularRoutes(e, deps)

	server := &EchoServer{
		echo: e,
		log:  deps.Logger,
	}

	deps.Lifecycle.Append(fx.Hook{
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

func setupModularRoutes(e *echo.Echo, deps ServerDeps) {
	// Register all routes - modules will handle their own middleware
	for _, registrar := range deps.RouteRegistrars {
		registrar.RegisterRoutes(e)
	}

	// For protected routes, we still create the groups but modules apply middleware individually
	api := e.Group("/api/v1")
	
	// Register protected routes - each module applies auth middleware as needed
	for _, registrar := range deps.ProtectedRegistrars {
		registrar.RegisterProtectedRoutes(api)
	}

	// Admin routes group - modules apply admin middleware as needed
	admin := api.Group("/admin")
	for _, registrar := range deps.AdminRegistrars {
		registrar.RegisterAdminRoutes(admin)
	}
}
