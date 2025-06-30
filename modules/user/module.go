package user

import (
	"fxserver/modules/user/repository"
	"fxserver/pkg/router"

	"go.uber.org/fx"
)

var Module = fx.Options(
	repository.Module,
	fx.Provide(
		NewService,
		NewHandler,
		NewAuthAdapter,
		fx.Annotate(
			NewRoutes,
			fx.As(new(router.RouteRegistrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),
)
