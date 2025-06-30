package user

import (
	"errors"
	"fxserver/pkg/jwt"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// PasswordVerifier is an interface to break circular dependency
type PasswordVerifier interface {
	VerifyUserPassword(email, password string) (UserInfo, error)
}

// UserInfo contains basic user information needed for authentication
type UserInfo struct {
	ID    int
	Email string
	Name  string
}

type Param struct {
	fx.In
	AccessTokenService  jwt.Service `name:"access_token"`
	RefreshTokenService jwt.Service `name:"refresh_token"`
	PasswordVerifier    PasswordVerifier
	Logger              *zap.Logger
}

type Service interface {
	Login(email, password string) (*LoginResponse, error)
	RefreshToken(refreshToken string) (*RefreshResponse, error)
	ValidateAccessToken(token string) (*jwt.Claims, error)
	ValidateRefreshToken(token string) (*jwt.Claims, error)
}

type service struct {
	accessTokenService  jwt.Service
	refreshTokenService jwt.Service
	passwordVerifier    PasswordVerifier
	logger              *zap.Logger
}

func NewService(p Param) Service {
	return &service{
		accessTokenService:  p.AccessTokenService,
		refreshTokenService: p.RefreshTokenService,
		passwordVerifier:    p.PasswordVerifier,
		logger:              p.Logger,
	}
}

func (s *service) Login(email, password string) (*LoginResponse, error) {
	// Use the password verifier interface to verify credentials
	userInfo, err := s.passwordVerifier.VerifyUserPassword(email, password)
	if err != nil {
		s.logger.Warn("Login failed", zap.String("email", email), zap.Error(err))
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.accessTokenService.GenerateToken(userInfo.ID, userInfo.Email)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := s.refreshTokenService.GenerateToken(userInfo.ID, userInfo.Email)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	response := &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	s.logger.Info("User logged in successfully", zap.Int("user_id", userInfo.ID), zap.String("email", email))
	return response, nil
}

func (s *service) RefreshToken(refreshToken string) (*RefreshResponse, error) {
	claims, err := s.refreshTokenService.ValidateToken(refreshToken)
	if err != nil {
		s.logger.Warn("Invalid refresh token", zap.Error(err))
		return nil, ErrInvalidRefreshToken
	}

	// Generate new access token
	accessToken, err := s.accessTokenService.GenerateToken(claims.UserID, claims.Email, claims.Role)
	if err != nil {
		s.logger.Error("Failed to generate new access token", zap.Error(err))
		return nil, err
	}

	response := &RefreshResponse{
		AccessToken: accessToken,
	}

	s.logger.Info("Token refreshed successfully", zap.Int("user_id", claims.UserID))
	return response, nil
}

func (s *service) ValidateAccessToken(token string) (*jwt.Claims, error) {
	return s.accessTokenService.ValidateToken(token)
}

func (s *service) ValidateRefreshToken(token string) (*jwt.Claims, error) {
	return s.refreshTokenService.ValidateToken(token)
}