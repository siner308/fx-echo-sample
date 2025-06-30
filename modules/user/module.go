package user

import (
	"fxserver/modules/user/repository"

	"go.uber.org/fx"
)

var Module = fx.Options(
	repository.Module,
	fx.Provide(
		NewService,
		NewHandler,
	),
)
