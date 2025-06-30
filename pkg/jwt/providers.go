package jwt

import (
	"os"
	"time"

	"go.uber.org/zap"
)

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

// NewAccessTokenService creates a JWT service for access tokens
func NewAccessTokenService(logger *zap.Logger) Service {
	config := Config{
		Secret:    getEnvOrDefault("ACCESS_TOKEN_SECRET", "access_secret_key_123"),
		ExpiresIn: parseExpirationTime("ACCESS_TOKEN_EXPIRES", "15m"),
		Issuer:    getEnvOrDefault("JWT_ISSUER", "fxserver"),
		TokenType: "access",
	}
	
	logger.Info("Creating access token service", 
		zap.String("expires_in", config.ExpiresIn.String()),
		zap.String("issuer", config.Issuer))
	
	return NewService(config, logger)
}

// NewRefreshTokenService creates a JWT service for refresh tokens
func NewRefreshTokenService(logger *zap.Logger) Service {
	config := Config{
		Secret:    getEnvOrDefault("REFRESH_TOKEN_SECRET", "refresh_secret_key_456"),
		ExpiresIn: parseExpirationTime("REFRESH_TOKEN_EXPIRES", "168h"), // 7 days
		Issuer:    getEnvOrDefault("JWT_ISSUER", "fxserver"),
		TokenType: "refresh",
	}
	
	logger.Info("Creating refresh token service", 
		zap.String("expires_in", config.ExpiresIn.String()),
		zap.String("issuer", config.Issuer))
	
	return NewService(config, logger)
}

// NewAdminTokenService creates a JWT service for admin tokens
func NewAdminTokenService(logger *zap.Logger) Service {
	config := Config{
		Secret:    getEnvOrDefault("ADMIN_TOKEN_SECRET", "admin_secret_key_789"),
		ExpiresIn: parseExpirationTime("ADMIN_TOKEN_EXPIRES", "1h"),
		Issuer:    getEnvOrDefault("JWT_ISSUER", "fxserver"),
		TokenType: "admin",
	}
	
	logger.Info("Creating admin token service", 
		zap.String("expires_in", config.ExpiresIn.String()),
		zap.String("issuer", config.Issuer))
	
	return NewService(config, logger)
}