package main

import (
	"fxserver/middleware"
	"fxserver/modules/auth"
	"fxserver/modules/coupon"
	"fxserver/modules/user"
	"fxserver/pkg/validator"
	"fxserver/server"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			zap.NewProduction,
			validator.New,
			middleware.NewLoggerMiddleware,
			middleware.NewErrorMiddleware,
			server.NewEchoServer,
		),
		auth.Module,
		user.Module,
		coupon.Module,
		fx.Invoke(func(s *server.EchoServer) {
			// Server will be started by lifecycle hooks
		}),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	).Run()
}
