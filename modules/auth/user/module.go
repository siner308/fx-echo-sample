package user

import (
	"fxserver/pkg/router"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewService,
		NewHandler,
		NewMiddleware,
		fx.Annotate(
			NewRoutes,
			fx.As(new(router.RouteRegistrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),
)