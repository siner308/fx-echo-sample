package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fxserver/pkg/jwt"
	"fxserver/pkg/keycloak"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials    = errors.New("invalid admin credentials")
	ErrUserNotFound         = errors.New("admin user not found")
	ErrNotAdminUser         = errors.New("user is not an admin")
	ErrAdminTokenUnavailable = errors.New("admin token service not available")
	ErrKeycloakUnavailable  = errors.New("keycloak service not available")
	ErrInvalidAuthCode      = errors.New("invalid authorization code")
	ErrTokenExchange        = errors.New("failed to exchange token")
)

type Param struct {
	fx.In
	AdminTokenService jwt.Service      `name:"admin_token,optional"`
	KeycloakClient    keycloak.Client
	Logger            *zap.Logger
}

type Service interface {
	GetKeycloakAuthURL() (string, error)
	HandleKeycloakCallback(ctx context.Context, code string) (*AdminLoginResponse, error)
	ValidateAdminToken(token string) (*jwt.Claims, error)
	ValidateKeycloakToken(ctx context.Context, accessToken string) (*keycloak.UserInfo, error)
}

type service struct {
	adminTokenService jwt.Service
	keycloakClient    keycloak.Client
	logger            *zap.Logger
}

func NewService(p Param) Service {
	return &service{
		adminTokenService: p.AdminTokenService,
		keycloakClient:    p.KeycloakClient,
		logger:            p.Logger,
	}
}

func (s *service) GetKeycloakAuthURL() (string, error) {
	if s.keycloakClient == nil {
		return "", ErrKeycloakUnavailable
	}

	// Generate random state for CSRF protection
	state, err := generateRandomState()
	if err != nil {
		s.logger.Error("Failed to generate state", zap.Error(err))
		return "", err
	}

	authURL := s.keycloakClient.GetAuthURL(state)
	s.logger.Info("Generated Keycloak auth URL", zap.String("state", state))
	
	return authURL, nil
}

func (s *service) HandleKeycloakCallback(ctx context.Context, code string) (*AdminLoginResponse, error) {
	if s.keycloakClient == nil {
		return nil, ErrKeycloakUnavailable
	}

	if code == "" {
		return nil, ErrInvalidAuthCode
	}

	// Exchange authorization code for tokens
	tokenResp, err := s.keycloakClient.ExchangeCodeForToken(ctx, code)
	if err != nil {
		s.logger.Error("Failed to exchange code for token", zap.Error(err))
		return nil, ErrTokenExchange
	}

	// Get user info from Keycloak
	userInfo, err := s.keycloakClient.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("Failed to get user info from Keycloak", zap.Error(err))
		return nil, err
	}

	// Check if user has admin role (customize this logic based on your Keycloak setup)
	if !hasAdminRole(userInfo) {
		s.logger.Warn("Non-admin user attempted admin login", 
			zap.String("email", userInfo.Email),
			zap.String("sub", userInfo.Sub))
		return nil, ErrNotAdminUser
	}

	// Generate our internal admin token if service is available
	var internalToken string
	var expiresIn int64

	if s.adminTokenService != nil {
		// Convert Keycloak user info to internal user ID (you might want to store this mapping)
		userID := hashSubToUserID(userInfo.Sub) // Simple hash for demo
		
		internalToken, err = s.adminTokenService.GenerateToken(userID, userInfo.Email, "admin")
		if err != nil {
			s.logger.Error("Failed to generate internal admin token", zap.Error(err))
			return nil, err
		}
		expiresIn = int64(s.adminTokenService.GetExpirationTime().Seconds())
	} else {
		// Use Keycloak token directly
		internalToken = tokenResp.AccessToken
		expiresIn = tokenResp.ExpiresIn
	}

	response := &AdminLoginResponse{
		AdminToken:     internalToken,
		ExpiresIn:      expiresIn,
		KeycloakToken:  tokenResp.AccessToken, // Include original Keycloak token
		RefreshToken:   tokenResp.RefreshToken,
		UserInfo:       userInfo,
	}

	s.logger.Info("Admin logged in successfully via Keycloak", 
		zap.String("sub", userInfo.Sub),
		zap.String("email", userInfo.Email))
	
	return response, nil
}

func (s *service) ValidateAdminToken(token string) (*jwt.Claims, error) {
	if s.adminTokenService == nil {
		return nil, ErrAdminTokenUnavailable
	}
	
	claims, err := s.adminTokenService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Verify admin role
	if claims.Role != "admin" {
		s.logger.Warn("Token validation failed: not an admin token", 
			zap.Int("user_id", claims.UserID),
			zap.String("role", claims.Role))
		return nil, ErrNotAdminUser
	}

	return claims, nil
}

func (s *service) ValidateKeycloakToken(ctx context.Context, accessToken string) (*keycloak.UserInfo, error) {
	if s.keycloakClient == nil {
		return nil, ErrKeycloakUnavailable
	}

	// Validate token with Keycloak
	introspection, err := s.keycloakClient.ValidateToken(ctx, accessToken)
	if err != nil {
		s.logger.Warn("Keycloak token validation failed", zap.Error(err))
		return nil, err
	}

	if !introspection.Active {
		return nil, errors.New("keycloak token is not active")
	}

	// Get user info
	userInfo, err := s.keycloakClient.GetUserInfo(ctx, accessToken)
	if err != nil {
		s.logger.Error("Failed to get user info from Keycloak", zap.Error(err))
		return nil, err
	}

	// Check admin role
	if !hasAdminRole(userInfo) {
		return nil, ErrNotAdminUser
	}

	return userInfo, nil
}

// Helper functions
func generateRandomState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hasAdminRole(userInfo *keycloak.UserInfo) bool {
	// Check if user has admin role in their roles
	for _, role := range userInfo.Roles {
		if role == "admin" || role == "realm-admin" || role == "admin-cli" {
			return true
		}
	}

	// Check in groups (alternative approach)
	for _, group := range userInfo.Groups {
		if group == "admin" || group == "administrators" {
			return true
		}
	}

	// For demo purposes, you could also check specific emails
	adminEmails := []string{"admin@example.com", "admin@localhost"}
	for _, adminEmail := range adminEmails {
		if userInfo.Email == adminEmail {
			return true
		}
	}

	return false
}

func hashSubToUserID(sub string) int {
	// Simple hash function to convert Keycloak sub to integer ID
	// In production, you'd want to maintain a proper mapping table
	hash := 0
	for _, char := range sub {
		hash = hash*31 + int(char)
	}
	
	// Ensure positive number and reasonable range
	if hash < 0 {
		hash = -hash
	}
	return (hash % 1000000) + 1000 // Range: 1000-1001000
}