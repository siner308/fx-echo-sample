package jwt

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TokenConfig holds required configuration for token types
type TokenConfig struct {
	ExpiresIn string
}

var tokenConfigs = map[string]TokenConfig{
	"access": {
		ExpiresIn: "15m",
	},
	"refresh": {
		ExpiresIn: "168h", // 7 days
	},
	"admin": {
		ExpiresIn: "1h",
	},
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Required environment variable not set: " + key)
	}
	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseExpirationTime(envKey, defaultValue string) time.Duration {
	value := getEnvOrDefault(envKey, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		// Fallback to default if parsing fails
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}

// newTokenService creates a JWT service with name-based configuration
func newTokenService(tokenType string, logger *zap.Logger) Service {
	config, exists := tokenConfigs[tokenType]
	if !exists {
		logger.Error("Unknown token type", zap.String("type", tokenType))
		panic("Unknown token type: " + tokenType)
	}

	// Generate environment variable names based on token type
	// ACCESS_TOKEN_SECRET, REFRESH_TOKEN_SECRET, ADMIN_TOKEN_SECRET
	secretEnvKey := strings.ToUpper(tokenType) + "_TOKEN_SECRET"
	expiresEnvKey := strings.ToUpper(tokenType) + "_TOKEN_EXPIRES"
	
	jwtConfig := Config{
		Secret:    getRequiredEnv(secretEnvKey),
		ExpiresIn: parseExpirationTime(expiresEnvKey, config.ExpiresIn),
		Issuer:    getRequiredEnv("JWT_ISSUER"),
		TokenType: tokenType,
	}
	
	logger.Info("Creating JWT token service", 
		zap.String("type", tokenType),
		zap.String("expires_in", jwtConfig.ExpiresIn.String()),
		zap.String("issuer", jwtConfig.Issuer),
		zap.String("secret_env_key", secretEnvKey),
		zap.String("expires_env_key", expiresEnvKey))
	
	return NewService(jwtConfig, logger)
}

// NewAccessTokenService creates a JWT service for access tokens
func NewAccessTokenService(logger *zap.Logger) Service {
	return newTokenService("access", logger)
}

// NewRefreshTokenService creates a JWT service for refresh tokens
func NewRefreshTokenService(logger *zap.Logger) Service {
	return newTokenService("refresh", logger)
}

// NewAdminTokenService creates a JWT service for admin tokens
func NewAdminTokenService(logger *zap.Logger) Service {
	return newTokenService("admin", logger)
}