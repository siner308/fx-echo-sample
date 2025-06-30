package auth

import (
	"fxserver/modules/auth/admin"
	"fxserver/modules/auth/user"
	"fxserver/pkg/jwt"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Provide JWT services with different names
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
	),

	// Include user and admin auth modules
	user.Module,
	admin.Module,
)
