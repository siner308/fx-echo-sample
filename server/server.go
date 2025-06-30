package server

import (
	"context"
	"errors"
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

type Param struct {
	fx.In
	Lifecycle        fx.Lifecycle
	Logger           *zap.Logger
	LoggerMiddleware *middleware.LoggerMiddleware
	ErrorMiddleware  *middleware.ErrorMiddleware
	RouteRegistrars  []router.RouteRegistrar `group:"routes"`
}

func NewEchoServer(p Param) *EchoServer {
	e := echo.New()

	// Set error handler
	e.HTTPErrorHandler = p.ErrorMiddleware.ErrorHandler()

	// Add middleware
	e.Use(p.LoggerMiddleware.LoggerMiddleware())

	// Setup routes using simplified registration
	setupRoutes(e, p.RouteRegistrars)

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
		if err := s.echo.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatal("Server failed to start", zap.Error(err))
		}
	}()
	return nil
}

func (s *EchoServer) Stop(ctx context.Context) error {
	s.log.Info("Stopping HTTP server")
	return s.echo.Shutdown(ctx)
}

func setupRoutes(e *echo.Echo, registrars []router.RouteRegistrar) {
	// Register all routes - each module handles its own middleware selection
	for _, r := range registrars {
		r.RegisterRoutes(e)
	}
}
