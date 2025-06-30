package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Client interface {
	GetAuthURL(state string) string
	ExchangeCodeForToken(ctx context.Context, code string) (*TokenResponse, error)
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)
	ValidateToken(ctx context.Context, accessToken string) (*TokenIntrospection, error)
}

type client struct {
	config Config
	logger *zap.Logger
	client *http.Client
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	IDToken          string `json:"id_token"`
	Scope            string `json:"scope"`
}

type UserInfo struct {
	Sub               string   `json:"sub"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	PreferredUsername string   `json:"preferred_username"`
	Name              string   `json:"name"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	Roles             []string `json:"realm_access.roles,omitempty"`
	Groups            []string `json:"groups,omitempty"`
}

type TokenIntrospection struct {
	Active    bool   `json:"active"`
	Sub       string `json:"sub"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	ClientID  string `json:"client_id"`
	TokenType string `json:"token_type"`
}

func NewClient(config Config, logger *zap.Logger) Client {
	return &client{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *client) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", c.config.ClientID)
	params.Add("redirect_uri", c.config.RedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("state", state)

	authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth?%s",
		c.config.BaseURL, c.config.Realm, params.Encode())

	c.logger.Info("Generated Keycloak auth URL", 
		zap.String("url", authURL),
		zap.String("state", state))

	return authURL
}

func (c *client) ExchangeCodeForToken(ctx context.Context, code string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		c.config.BaseURL, c.config.Realm)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", c.config.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Failed to exchange code for token", zap.Error(err))
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Token exchange failed", 
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	c.logger.Info("Successfully exchanged code for token")
	return &tokenResp, nil
}

func (c *client) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	userInfoURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo",
		c.config.BaseURL, c.config.Realm)

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Failed to get user info", zap.Error(err))
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read userinfo response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("UserInfo request failed", 
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal userinfo response: %w", err)
	}

	c.logger.Info("Successfully retrieved user info", zap.String("sub", userInfo.Sub))
	return &userInfo, nil
}

func (c *client) ValidateToken(ctx context.Context, accessToken string) (*TokenIntrospection, error) {
	introspectURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token/introspect",
		c.config.BaseURL, c.config.Realm)

	data := url.Values{}
	data.Set("token", accessToken)
	data.Set("client_id", c.config.ClientID)
	data.Set("client_secret", c.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Failed to validate token", zap.Error(err))
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read introspection response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Token validation failed", 
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("token validation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var introspection TokenIntrospection
	if err := json.Unmarshal(body, &introspection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal introspection response: %w", err)
	}

	if !introspection.Active {
		return nil, fmt.Errorf("token is not active")
	}

	c.logger.Info("Token validation successful", zap.String("sub", introspection.Sub))
	return &introspection, nil
}