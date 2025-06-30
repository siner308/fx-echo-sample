package main

import (
	"fxserver/middleware"
	"fxserver/modules/auth"
	"fxserver/modules/coupon"
	"fxserver/modules/item"
	"fxserver/modules/payment"
	"fxserver/modules/reward"
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
		item.Module,     // 기본 아이템 시스템
		payment.Module,  // 결제 처리 (item 의존)
		reward.Module,   // 통합 보상 시스템 (item, payment 의존)
		user.Module,
		coupon.Module,   // 쿠폰 시스템 (reward 의존하여 아이템 지급)
		fx.Invoke(func(s *server.EchoServer) {
			// Server will be started by lifecycle hooks
		}),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	).Run()
}
