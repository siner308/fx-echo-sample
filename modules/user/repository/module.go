package repository

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewMemoryUserRepository,
			fx.As(new(UserRepository)),
		),
	),
)