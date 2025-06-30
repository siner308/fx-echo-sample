package user

import (
	"errors"
	"fxserver/modules/user/entity"
	"fxserver/modules/user/repository"
	"fxserver/pkg/jwt"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type Param struct {
	fx.In
	AccessTokenService  jwt.Service `name:"access_token"`
	RefreshTokenService jwt.Service `name:"refresh_token"`
	UserRepository      repository.UserRepository
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
	userRepository      repository.UserRepository
	logger              *zap.Logger
}

func NewService(p Param) Service {
	return &service{
		accessTokenService:  p.AccessTokenService,
		refreshTokenService: p.RefreshTokenService,
		userRepository:      p.UserRepository,
		logger:              p.Logger,
	}
}

func (s *service) Login(email, password string) (*LoginResponse, error) {
	// In a real implementation, you would:
	// 1. Get user by email
	// 2. Verify password hash
	// For this demo, we'll simulate with existing user data
	
	users, err := s.userRepository.GetAll()
	if err != nil {
		s.logger.Error("Failed to list users for login", zap.Error(err))
		return nil, err
	}

	var user *entity.User
	for _, u := range users {
		if u.Email == email {
			user = u
			break
		}
	}

	if user == nil {
		s.logger.Warn("Login attempt with non-existent email", zap.String("email", email))
		return nil, ErrUserNotFound
	}

	// In real implementation, verify password hash
	// For demo, we'll just check if password is not empty
	if password == "" {
		s.logger.Warn("Login attempt with empty password", zap.String("email", email))
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.accessTokenService.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := s.refreshTokenService.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	response := &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenService.GetExpirationTime().Seconds()),
		User:         user.ToResponse(),
	}

	s.logger.Info("User logged in successfully", zap.Int("user_id", user.ID), zap.String("email", email))
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
		ExpiresIn:   int64(s.accessTokenService.GetExpirationTime().Seconds()),
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