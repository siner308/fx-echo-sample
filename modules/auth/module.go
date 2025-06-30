package auth

import (
	"fxserver/pkg/jwt"
	"fxserver/pkg/router"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Provide JWT services with different names based on environment variables
	fx.Provide(
		fx.Annotate(
			jwt.NewAccessTokenService,
			fx.As(new(jwt.Service)),
			fx.ResultTags(`name:"access_token"`),
		),
		fx.Annotate(
			jwt.NewRefreshTokenService,
			fx.As(new(jwt.Service)),
			fx.ResultTags(`name:"refresh_token"`),
		),
		fx.Annotate(
			jwt.NewAdminTokenService,
			fx.As(new(jwt.Service)),
			fx.ResultTags(`name:"admin_token"`),
		),
	),
	
	// Provide auth services
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