package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Service interface {
	GenerateToken(userID int, email string, role ...string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	GetExpirationTime() time.Duration
}

type Config struct {
	Secret     string
	ExpiresIn  time.Duration
	Issuer     string
	TokenType  string // "access", "refresh", "admin"
}