package user

import (
	"errors"
	"testing"
	"time"

	"fxserver/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock JWT Service
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(userID int, email string, roles ...string) (string, error) {
	// Convert variadic args to interface slice for mock
	var mockArgs []interface{}
	mockArgs = append(mockArgs, userID, email)
	for _, role := range roles {
		mockArgs = append(mockArgs, role)
	}
	args := m.Called(mockArgs...)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

func (m *MockJWTService) GetExpirationTime() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

// Mock Password Verifier
type MockPasswordVerifier struct {
	mock.Mock
}

func (m *MockPasswordVerifier) VerifyUserPassword(email, password string) (UserInfo, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return UserInfo{}, args.Error(1)
	}
	return args.Get(0).(UserInfo), args.Error(1)
}

func setupAuthService(accessTokenService, refreshTokenService *MockJWTService, passwordVerifier *MockPasswordVerifier) Service {
	logger := zap.NewNop()
	return &service{
		accessTokenService:  accessTokenService,
		refreshTokenService: refreshTokenService,
		passwordVerifier:    passwordVerifier,
		logger:              logger,
	}
}

func TestLogin(t *testing.T) {
	testUser := UserInfo{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
	}

	tests := []struct {
		name      string
		email     string
		password  string
		setupMock func(*MockJWTService, *MockJWTService, *MockPasswordVerifier)
		wantErr   bool
		wantUser  bool
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			setupMock: func(access, refresh *MockJWTService, verifier *MockPasswordVerifier) {
				verifier.On("VerifyUserPassword", "test@example.com", "password123").Return(testUser, nil)
				access.On("GenerateToken", 1, "test@example.com").Return("access_token", nil)
				refresh.On("GenerateToken", 1, "test@example.com").Return("refresh_token", nil)
			},
			wantErr:  false,
			wantUser: true,
		},
		{
			name:     "invalid credentials",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMock: func(access, refresh *MockJWTService, verifier *MockPasswordVerifier) {
				verifier.On("VerifyUserPassword", "test@example.com", "wrongpassword").Return(UserInfo{}, errors.New("invalid credentials"))
			},
			wantErr:  true,
			wantUser: false,
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMock: func(access, refresh *MockJWTService, verifier *MockPasswordVerifier) {
				verifier.On("VerifyUserPassword", "nonexistent@example.com", "password123").Return(UserInfo{}, errors.New("user not found"))
			},
			wantErr:  true,
			wantUser: false,
		},
		{
			name:     "access token generation failed",
			email:    "test@example.com",
			password: "password123",
			setupMock: func(access, refresh *MockJWTService, verifier *MockPasswordVerifier) {
				verifier.On("VerifyUserPassword", "test@example.com", "password123").Return(testUser, nil)
				access.On("GenerateToken", 1, "test@example.com").Return("", errors.New("token generation failed"))
			},
			wantErr:  true,
			wantUser: false,
		},
		{
			name:     "refresh token generation failed",
			email:    "test@example.com",
			password: "password123",
			setupMock: func(access, refresh *MockJWTService, verifier *MockPasswordVerifier) {
				verifier.On("VerifyUserPassword", "test@example.com", "password123").Return(testUser, nil)
				access.On("GenerateToken", 1, "test@example.com").Return("access_token", nil)
				refresh.On("GenerateToken", 1, "test@example.com").Return("", errors.New("refresh token generation failed"))
			},
			wantErr:  true,
			wantUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessTokenService := new(MockJWTService)
			refreshTokenService := new(MockJWTService)
			passwordVerifier := new(MockPasswordVerifier)

			tt.setupMock(accessTokenService, refreshTokenService, passwordVerifier)

			service := setupAuthService(accessTokenService, refreshTokenService, passwordVerifier)

			response, err := service.Login(tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, "access_token", response.AccessToken)
				assert.Equal(t, "refresh_token", response.RefreshToken)
			}

			accessTokenService.AssertExpectations(t)
			refreshTokenService.AssertExpectations(t)
			passwordVerifier.AssertExpectations(t)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	validClaims := &jwt.Claims{
		UserID: 1,
		Email:  "test@example.com",
		Role:   "user",
	}

	tests := []struct {
		name         string
		refreshToken string
		setupMock    func(*MockJWTService, *MockJWTService)
		wantErr      bool
	}{
		{
			name:         "successful token refresh",
			refreshToken: "valid_refresh_token",
			setupMock: func(access, refresh *MockJWTService) {
				refresh.On("ValidateToken", "valid_refresh_token").Return(validClaims, nil)
				access.On("GenerateToken", 1, "test@example.com", "user").Return("new_access_token", nil)
			},
			wantErr: false,
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid_refresh_token",
			setupMock: func(access, refresh *MockJWTService) {
				refresh.On("ValidateToken", "invalid_refresh_token").Return(nil, errors.New("invalid token"))
			},
			wantErr: true,
		},
		{
			name:         "access token generation failed",
			refreshToken: "valid_refresh_token",
			setupMock: func(access, refresh *MockJWTService) {
				refresh.On("ValidateToken", "valid_refresh_token").Return(validClaims, nil)
				access.On("GenerateToken", 1, "test@example.com", "user").Return("", errors.New("token generation failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessTokenService := new(MockJWTService)
			refreshTokenService := new(MockJWTService)
			passwordVerifier := new(MockPasswordVerifier)

			tt.setupMock(accessTokenService, refreshTokenService)

			service := setupAuthService(accessTokenService, refreshTokenService, passwordVerifier)

			response, err := service.RefreshToken(tt.refreshToken)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, "new_access_token", response.AccessToken)
			}

			accessTokenService.AssertExpectations(t)
			refreshTokenService.AssertExpectations(t)
		})
	}
}

func TestValidateAccessToken(t *testing.T) {
	validClaims := &jwt.Claims{
		UserID: 1,
		Email:  "test@example.com",
		Role:   "user",
	}

	tests := []struct {
		name      string
		token     string
		setupMock func(*MockJWTService)
		wantErr   bool
	}{
		{
			name:  "valid access token",
			token: "valid_access_token",
			setupMock: func(access *MockJWTService) {
				access.On("ValidateToken", "valid_access_token").Return(validClaims, nil)
			},
			wantErr: false,
		},
		{
			name:  "invalid access token",
			token: "invalid_access_token",
			setupMock: func(access *MockJWTService) {
				access.On("ValidateToken", "invalid_access_token").Return(nil, errors.New("invalid token"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessTokenService := new(MockJWTService)
			refreshTokenService := new(MockJWTService)
			passwordVerifier := new(MockPasswordVerifier)

			tt.setupMock(accessTokenService)

			service := setupAuthService(accessTokenService, refreshTokenService, passwordVerifier)

			claims, err := service.ValidateAccessToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, validClaims.UserID, claims.UserID)
				assert.Equal(t, validClaims.Email, claims.Email)
				assert.Equal(t, validClaims.Role, claims.Role)
			}

			accessTokenService.AssertExpectations(t)
		})
	}
}

func TestValidateRefreshToken(t *testing.T) {
	validClaims := &jwt.Claims{
		UserID: 1,
		Email:  "test@example.com",
		Role:   "user",
	}

	tests := []struct {
		name      string
		token     string
		setupMock func(*MockJWTService)
		wantErr   bool
	}{
		{
			name:  "valid refresh token",
			token: "valid_refresh_token",
			setupMock: func(refresh *MockJWTService) {
				refresh.On("ValidateToken", "valid_refresh_token").Return(validClaims, nil)
			},
			wantErr: false,
		},
		{
			name:  "invalid refresh token",
			token: "invalid_refresh_token",
			setupMock: func(refresh *MockJWTService) {
				refresh.On("ValidateToken", "invalid_refresh_token").Return(nil, errors.New("invalid token"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessTokenService := new(MockJWTService)
			refreshTokenService := new(MockJWTService)
			passwordVerifier := new(MockPasswordVerifier)

			tt.setupMock(refreshTokenService)

			service := setupAuthService(accessTokenService, refreshTokenService, passwordVerifier)

			claims, err := service.ValidateRefreshToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, validClaims.UserID, claims.UserID)
				assert.Equal(t, validClaims.Email, claims.Email)
				assert.Equal(t, validClaims.Role, claims.Role)
			}

			refreshTokenService.AssertExpectations(t)
		})
	}
}

// Integration test for login error types
func TestLoginErrorTypes(t *testing.T) {
	accessTokenService := new(MockJWTService)
	refreshTokenService := new(MockJWTService)
	passwordVerifier := new(MockPasswordVerifier)

	passwordVerifier.On("VerifyUserPassword", "test@example.com", "wrongpassword").Return(UserInfo{}, errors.New("invalid credentials"))

	service := setupAuthService(accessTokenService, refreshTokenService, passwordVerifier)

	response, err := service.Login("test@example.com", "wrongpassword")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, response)

	passwordVerifier.AssertExpectations(t)
}

func TestRefreshTokenErrorTypes(t *testing.T) {
	accessTokenService := new(MockJWTService)
	refreshTokenService := new(MockJWTService)
	passwordVerifier := new(MockPasswordVerifier)

	refreshTokenService.On("ValidateToken", "invalid_token").Return(nil, errors.New("token expired"))

	service := setupAuthService(accessTokenService, refreshTokenService, passwordVerifier)

	response, err := service.RefreshToken("invalid_token")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRefreshToken, err)
	assert.Nil(t, response)

	refreshTokenService.AssertExpectations(t)
}

// Test concurrent login attempts (race condition test)
func TestConcurrentLogin(t *testing.T) {
	testUser := UserInfo{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
	}

	accessTokenService := new(MockJWTService)
	refreshTokenService := new(MockJWTService)
	passwordVerifier := new(MockPasswordVerifier)

	// Setup mocks for multiple concurrent calls
	passwordVerifier.On("VerifyUserPassword", "test@example.com", "password123").Return(testUser, nil)
	accessTokenService.On("GenerateToken", 1, "test@example.com").Return("access_token", nil)
	refreshTokenService.On("GenerateToken", 1, "test@example.com").Return("refresh_token", nil)
	accessTokenService.On("GetExpirationTime").Return(time.Hour)

	service := setupAuthService(accessTokenService, refreshTokenService, passwordVerifier)

	// Run concurrent login attempts
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := service.Login("test@example.com", "password123")
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for any errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent login failed: %v", err)
	}

	// Note: We can't easily assert call counts with testify/mock in concurrent scenarios
	// but this test ensures no race conditions or panics occur
}