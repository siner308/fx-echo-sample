package keycloak

import (
	"os"

	"go.uber.org/zap"
)

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Required environment variable not set: " + key)
	}
	return value
}

// NewKeycloakClient creates a Keycloak client from environment variables
func NewKeycloakClient(logger *zap.Logger) Client {
	config := Config{
		BaseURL:      getRequiredEnv("KEYCLOAK_BASE_URL"),
		Realm:        getRequiredEnv("KEYCLOAK_REALM"),
		ClientID:     getRequiredEnv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: getRequiredEnv("KEYCLOAK_CLIENT_SECRET"),
		RedirectURL:  getRequiredEnv("KEYCLOAK_REDIRECT_URL"),
	}

	logger.Info("Creating Keycloak client",
		zap.String("base_url", config.BaseURL),
		zap.String("realm", config.Realm),
		zap.String("client_id", config.ClientID),
		zap.String("redirect_url", config.RedirectURL))

	return NewClient(config, logger)
}
