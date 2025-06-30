package user

import (
	"testing"
	"time"

	"fxserver/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// 실제 JWT 서비스를 사용한 통합 테스트
func TestLoginWithRealJWTService(t *testing.T) {
	// JWT 서비스 설정
	accessConfig := jwt.Config{
		Secret:    "test-access-secret-key-32-bytes",
		ExpiresIn: time.Hour,
		Issuer:    "fx-echo-sample",
		TokenType: "access",
	}
	
	refreshConfig := jwt.Config{
		Secret:    "test-refresh-secret-key-32-bytes",
		ExpiresIn: 24 * time.Hour,
		Issuer:    "fx-echo-sample",
		TokenType: "refresh",
	}

	logger := zap.NewNop()
	accessTokenService := jwt.NewService(accessConfig, logger)
	refreshTokenService := jwt.NewService(refreshConfig, logger)

	// Mock password verifier
	passwordVerifier := new(MockPasswordVerifier)
	testUser := UserInfo{
		ID:    123,
		Email: "test@example.com",
		Name:  "Test User",
	}
	passwordVerifier.On("VerifyUserPassword", "test@example.com", "password123").Return(testUser, nil)

	// 서비스 생성
	authService := &service{
		accessTokenService:  accessTokenService,
		refreshTokenService: refreshTokenService,
		passwordVerifier:    passwordVerifier,
		logger:              logger,
	}

	// 로그인 테스트
	response, err := authService.Login("test@example.com", "password123")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)

	// Access Token 디코드 및 검증
	accessClaims, err := accessTokenService.ValidateToken(response.AccessToken)
	assert.NoError(t, err)
	assert.NotNil(t, accessClaims)
	assert.Equal(t, testUser.ID, accessClaims.UserID)
	assert.Equal(t, testUser.Email, accessClaims.Email)
	assert.Equal(t, testUser.Email, accessClaims.Subject)
	assert.Equal(t, "fx-echo-sample", accessClaims.Issuer)
	assert.Equal(t, "access", accessClaims.ID) // 토큰 타입

	// Refresh Token 디코드 및 검증
	refreshClaims, err := refreshTokenService.ValidateToken(response.RefreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, refreshClaims)
	assert.Equal(t, testUser.ID, refreshClaims.UserID)
	assert.Equal(t, testUser.Email, refreshClaims.Email)
	assert.Equal(t, testUser.Email, refreshClaims.Subject)
	assert.Equal(t, "fx-echo-sample", refreshClaims.Issuer)
	assert.Equal(t, "refresh", refreshClaims.ID) // 토큰 타입

	// 토큰 만료 시간 검증
	now := time.Now()
	assert.True(t, accessClaims.ExpiresAt.Time.After(now))
	assert.True(t, accessClaims.ExpiresAt.Time.Before(now.Add(2*time.Hour))) // 1시간 + 여유
	assert.True(t, refreshClaims.ExpiresAt.Time.After(now.Add(23*time.Hour))) // 24시간 - 여유
	assert.True(t, refreshClaims.ExpiresAt.Time.Before(now.Add(25*time.Hour))) // 24시간 + 여유

	passwordVerifier.AssertExpectations(t)
}

func TestRefreshTokenWithRealJWTService(t *testing.T) {
	// JWT 서비스 설정
	accessConfig := jwt.Config{
		Secret:    "test-access-secret-key-32-bytes",
		ExpiresIn: time.Hour,
		Issuer:    "fx-echo-sample",
		TokenType: "access",
	}
	
	refreshConfig := jwt.Config{
		Secret:    "test-refresh-secret-key-32-bytes",
		ExpiresIn: 24 * time.Hour,
		Issuer:    "fx-echo-sample",
		TokenType: "refresh",
	}

	logger := zap.NewNop()
	accessTokenService := jwt.NewService(accessConfig, logger)
	refreshTokenService := jwt.NewService(refreshConfig, logger)

	// 서비스 생성
	authService := &service{
		accessTokenService:  accessTokenService,
		refreshTokenService: refreshTokenService,
		passwordVerifier:    nil, // refresh에서는 사용하지 않음
		logger:              logger,
	}

	// 먼저 refresh token 생성
	testUserID := 456
	testEmail := "refresh@example.com"
	testRole := "user"
	
	refreshToken, err := refreshTokenService.GenerateToken(testUserID, testEmail, testRole)
	assert.NoError(t, err)

	// Refresh Token으로 새 Access Token 생성
	response, err := authService.RefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.AccessToken)

	// 새로 생성된 Access Token 검증
	newAccessClaims, err := accessTokenService.ValidateToken(response.AccessToken)
	assert.NoError(t, err)
	assert.NotNil(t, newAccessClaims)
	assert.Equal(t, testUserID, newAccessClaims.UserID)
	assert.Equal(t, testEmail, newAccessClaims.Email)
	assert.Equal(t, testRole, newAccessClaims.Role)
	assert.Equal(t, "fx-echo-sample", newAccessClaims.Issuer)
	assert.Equal(t, "access", newAccessClaims.ID)

	// 토큰 만료 시간 검증
	now := time.Now()
	assert.True(t, newAccessClaims.ExpiresAt.Time.After(now))
	assert.True(t, newAccessClaims.ExpiresAt.Time.Before(now.Add(2*time.Hour)))
}

func TestJWTTokenCrossValidation(t *testing.T) {
	// 다른 타입의 토큰이 검증되지 않는지 테스트
	accessConfig := jwt.Config{
		Secret:    "test-access-secret-key-32-bytes",
		ExpiresIn: time.Hour,
		Issuer:    "fx-echo-sample",
		TokenType: "access",
	}
	
	refreshConfig := jwt.Config{
		Secret:    "test-refresh-secret-key-32-bytes",
		ExpiresIn: 24 * time.Hour,
		Issuer:    "fx-echo-sample",
		TokenType: "refresh",
	}

	logger := zap.NewNop()
	accessTokenService := jwt.NewService(accessConfig, logger)
	refreshTokenService := jwt.NewService(refreshConfig, logger)

	// Access Token 생성
	accessToken, err := accessTokenService.GenerateToken(1, "test@example.com")
	assert.NoError(t, err)

	// Refresh Token 생성
	refreshToken, err := refreshTokenService.GenerateToken(1, "test@example.com")
	assert.NoError(t, err)

	// Access Token을 Refresh Token 서비스로 검증 시도 (실패해야 함)
	_, err = refreshTokenService.ValidateToken(accessToken)
	assert.Error(t, err, "Access token should not be valid for refresh token service")

	// Refresh Token을 Access Token 서비스로 검증 시도 (실패해야 함)
	_, err = accessTokenService.ValidateToken(refreshToken)
	assert.Error(t, err, "Refresh token should not be valid for access token service")

	// 올바른 검증은 성공해야 함
	accessClaims, err := accessTokenService.ValidateToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, "access", accessClaims.ID)

	refreshClaims, err := refreshTokenService.ValidateToken(refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, "refresh", refreshClaims.ID)
}

func TestInvalidTokens(t *testing.T) {
	config := jwt.Config{
		Secret:    "test-secret-key-32-bytes-long",
		ExpiresIn: time.Hour,
		Issuer:    "fx-echo-sample",
		TokenType: "access",
	}

	logger := zap.NewNop()
	tokenService := jwt.NewService(config, logger)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "invalid format",
			token: "invalid-token",
		},
		{
			name:  "fake jwt",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		},
		{
			name:  "malformed jwt",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := tokenService.ValidateToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}