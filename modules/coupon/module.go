package coupon

import (
	"fxserver/modules/coupon/repository"
	"go.uber.org/fx"
)

var Module = fx.Options(
	repository.Module,
	fx.Provide(
		NewService,
		NewHandler,
	),
)