package repository

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewMemoryCouponRepository,
			fx.As(new(CouponRepository)),
		),
	),
)