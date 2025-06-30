package server

import (
	"context"
	"net/http"

	"fxserver/middleware"
	"fxserver/pkg/router"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type EchoServer struct {
	echo *echo.Echo
	log  *zap.Logger
}

type ServerParam struct {
	fx.In
	Lifecycle        fx.Lifecycle
	Logger           *zap.Logger
	LoggerMiddleware *middleware.LoggerMiddleware
	ErrorMiddleware  *middleware.ErrorMiddleware
	RouteRegistrars  []router.RouteRegistrar `group:"routes"`
}

func NewEchoServer(p ServerParam) *EchoServer {
	e := echo.New()

	// Set error handler
	e.HTTPErrorHandler = p.ErrorMiddleware.ErrorHandler()

	// Add middleware
	e.Use(p.LoggerMiddleware.LoggerMiddleware())

	// Setup routes using simplified registration
	setupRoutes(e, p)

	server := &EchoServer{
		echo: e,
		log:  p.Logger,
	}

	p.Lifecycle.Append(fx.Hook{
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

func setupRoutes(e *echo.Echo, p ServerParam) {
	// Register all routes - each module handles its own middleware selection
	for _, registrar := range p.RouteRegistrars {
		registrar.RegisterRoutes(e)
	}
}
