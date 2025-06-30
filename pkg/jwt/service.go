package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
	ErrInvalidClaims = errors.New("invalid claims")
)

type service struct {
	config Config
	logger *zap.Logger
}

func NewService(config Config, logger *zap.Logger) Service {
	return &service{
		config: config,
		logger: logger,
	}
}

func (s *service) GenerateToken(userID int, email string, role ...string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.ExpiresIn)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   email,
			ID:        s.config.TokenType,
		},
	}

	if len(role) > 0 && role[0] != "" {
		claims.Role = role[0]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		s.logger.Error("Failed to sign token", 
			zap.Error(err), 
			zap.String("token_type", s.config.TokenType),
			zap.Int("user_id", userID))
		return "", err
	}

	s.logger.Info("Token generated successfully", 
		zap.String("token_type", s.config.TokenType),
		zap.Int("user_id", userID),
		zap.Time("expires_at", expiresAt))

	return tokenString, nil
}

func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		s.logger.Warn("Failed to parse token", 
			zap.Error(err),
			zap.String("token_type", s.config.TokenType))
		
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		s.logger.Warn("Invalid token claims", 
			zap.String("token_type", s.config.TokenType))
		return nil, ErrInvalidClaims
	}

	// Verify token type matches service configuration
	if claims.ID != s.config.TokenType {
		s.logger.Warn("Token type mismatch", 
			zap.String("expected", s.config.TokenType),
			zap.String("actual", claims.ID))
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *service) GetExpirationTime() time.Duration {
	return s.config.ExpiresIn
}