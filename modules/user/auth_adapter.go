package user

import (
	userAuth "fxserver/modules/auth/user"
	"go.uber.org/fx"
)

// AuthAdapter implements the PasswordVerifier interface for auth module
type AuthAdapter struct {
	userService Service
}

type AuthAdapterParam struct {
	fx.In
	UserService Service
}

func NewAuthAdapter(p AuthAdapterParam) userAuth.PasswordVerifier {
	return &AuthAdapter{
		userService: p.UserService,
	}
}

// VerifyUserPassword implements userAuth.PasswordVerifier interface
func (a *AuthAdapter) VerifyUserPassword(email, password string) (userAuth.UserInfo, error) {
	user, err := a.userService.VerifyUserPassword(email, password)
	if err != nil {
		return userAuth.UserInfo{}, err
	}

	// Convert entity.User to userAuth.UserInfo
	return userAuth.UserInfo{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}, nil
}
