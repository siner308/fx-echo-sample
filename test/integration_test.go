package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fxserver/modules/auth/user"
	itemModule "fxserver/modules/item"
	itemEntity "fxserver/modules/item/entity"
	userModule "fxserver/modules/user"
	"fxserver/pkg/jwt"
	"fxserver/pkg/security"
	"fxserver/pkg/validator"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type IntegrationTestSuite struct {
	suite.Suite
	echo           *echo.Echo
	userService    userModule.Service
	authService    user.Service
	itemService    itemModule.Service
	jwtService     jwt.Service
	testUser       *userModule.CreateUserRequest
	testAuthToken  string
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Setup Echo
	suite.echo = echo.New()
	logger := zap.NewNop()
	validator := validator.New()

	// Setup JWT service for testing
	jwtConfig := &jwt.Config{
		Secret:     "test-secret-key-for-integration-tests",
		ExpiresIn:  time.Hour,
		TokenType:  "access",
	}
	suite.jwtService = jwt.NewService(jwtConfig, logger)

	// Setup services with in-memory repositories
	// Note: In real integration tests, you'd use actual database connections
	suite.userService = setupInMemoryUserService(logger)
	suite.authService = setupInMemoryAuthService(suite.jwtService, suite.userService, logger)
	suite.itemService = setupInMemoryItemService(logger)

	// Setup test user
	suite.testUser = &userModule.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Password: "password123",
	}

	// Create test user and generate auth token
	createdUser, err := suite.userService.CreateUser(*suite.testUser)
	suite.Require().NoError(err)
	suite.Require().NotNil(createdUser)

	// Generate auth token
	token, err := suite.jwtService.GenerateToken(createdUser.ID, createdUser.Email)
	suite.Require().NoError(err)
	suite.testAuthToken = token
}

func (suite *IntegrationTestSuite) TestUserRegistrationAndLogin() {
	t := suite.T()

	// Test user registration
	registerRequest := userModule.CreateUserRequest{
		Name:     "New User",
		Email:    "newuser@example.com",
		Age:      30,
		Password: "newpassword123",
	}

	user, err := suite.userService.CreateUser(registerRequest)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, registerRequest.Name, user.Name)
	assert.Equal(t, registerRequest.Email, user.Email)
	assert.Equal(t, registerRequest.Age, user.Age)

	// Verify password is hashed
	assert.NotEqual(t, registerRequest.Password, user.Password)
	isValid, err := security.VerifyPassword(registerRequest.Password, user.Password)
	assert.NoError(t, err)
	assert.True(t, isValid)

	// Test login with correct credentials
	loginResponse, err := suite.authService.Login(registerRequest.Email, registerRequest.Password)
	assert.NoError(t, err)
	assert.NotNil(t, loginResponse)
	assert.NotEmpty(t, loginResponse.AccessToken)
	assert.NotEmpty(t, loginResponse.RefreshToken)
	assert.Equal(t, user.ID, loginResponse.User.ID)
	assert.Equal(t, user.Email, loginResponse.User.Email)

	// Test login with wrong password
	_, err = suite.authService.Login(registerRequest.Email, "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, user.ErrInvalidCredentials, err)

	// Test login with non-existent user
	_, err = suite.authService.Login("nonexistent@example.com", "password")
	assert.Error(t, err)
}

func (suite *IntegrationTestSuite) TestTokenRefresh() {
	t := suite.T()

	// Login to get tokens
	loginResponse, err := suite.authService.Login(suite.testUser.Email, suite.testUser.Password)
	assert.NoError(t, err)
	assert.NotNil(t, loginResponse)

	// Test token refresh
	refreshResponse, err := suite.authService.RefreshToken(loginResponse.RefreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, refreshResponse)
	assert.NotEmpty(t, refreshResponse.AccessToken)
	assert.NotEqual(t, loginResponse.AccessToken, refreshResponse.AccessToken) // New token should be different

	// Test refresh with invalid token
	_, err = suite.authService.RefreshToken("invalid_token")
	assert.Error(t, err)
	assert.Equal(t, user.ErrInvalidRefreshToken, err)
}

func (suite *IntegrationTestSuite) TestItemManagement() {
	t := suite.T()

	// Test item creation
	createRequest := itemModule.CreateItemRequest{
		Name:        "Integration Test Sword",
		Description: "A sword for testing",
		Type:        string(itemEntity.ItemTypeEquipment),
		Rarity:      string(itemEntity.RarityRare),
		IsActive:    true,
	}

	item, err := suite.itemService.CreateItem(createRequest)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, createRequest.Name, item.Name)
	assert.Equal(t, itemEntity.ItemTypeEquipment, item.Type)
	assert.Equal(t, itemEntity.RarityRare, item.Rarity)

	// Test item retrieval
	retrievedItem, err := suite.itemService.GetItem(item.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedItem)
	assert.Equal(t, item.ID, retrievedItem.ID)

	// Test item list
	items, err := suite.itemService.GetItems()
	assert.NoError(t, err)
	assert.NotNil(t, items)
	assert.GreaterOrEqual(t, len(items), 1)

	// Test items by type
	equipmentItems, err := suite.itemService.GetItemsByType(itemEntity.ItemTypeEquipment)
	assert.NoError(t, err)
	assert.NotNil(t, equipmentItems)

	// Test inventory addition
	userID := 1
	err = suite.itemService.AddToInventory(userID, item.ID, 5, "test")
	assert.NoError(t, err)

	// Test inventory retrieval
	inventory, err := suite.itemService.GetUserInventory(userID)
	assert.NoError(t, err)
	assert.NotNil(t, inventory)
	assert.Equal(t, userID, inventory.UserID)
	assert.GreaterOrEqual(t, inventory.TotalItems, 1)
}

func (suite *IntegrationTestSuite) TestUserProfileUpdate() {
	t := suite.T()

	// Get the test user
	users, err := suite.userService.ListUsers()
	assert.NoError(t, err)
	assert.NotEmpty(t, users)

	testUser := users[0]

	// Test profile update
	updateRequest := userModule.UpdateUserRequest{
		Name:     "Updated Name",
		Email:    "updated@example.com",
		Age:      26,
		Password: "newpassword456",
	}

	updatedUser, err := suite.userService.UpdateUser(testUser.ID, updateRequest)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, updateRequest.Name, updatedUser.Name)
	assert.Equal(t, updateRequest.Email, updatedUser.Email)
	assert.Equal(t, updateRequest.Age, updatedUser.Age)

	// Verify password was updated
	assert.NotEqual(t, testUser.Password, updatedUser.Password)
	isValid, err := security.VerifyPassword(updateRequest.Password, updatedUser.Password)
	assert.NoError(t, err)
	assert.True(t, isValid)

	// Test login with new password
	_, err = suite.authService.Login(updateRequest.Email, updateRequest.Password)
	assert.NoError(t, err)

	// Test login with old password should fail
	_, err = suite.authService.Login(updateRequest.Email, suite.testUser.Password)
	assert.Error(t, err)
}

func (suite *IntegrationTestSuite) TestRewardItemsFlow() {
	t := suite.T()

	userID := 1

	// Create test items
	sword, err := suite.itemService.CreateItem(itemModule.CreateItemRequest{
		Name:        "Test Sword",
		Description: "A test sword",
		Type:        string(itemEntity.ItemTypeEquipment),
		Rarity:      string(itemEntity.RarityRare),
		IsActive:    true,
	})
	assert.NoError(t, err)

	gold, err := suite.itemService.CreateItem(itemModule.CreateItemRequest{
		Name:        "Gold",
		Description: "Currency",
		Type:        string(itemEntity.ItemTypeCurrency),
		Rarity:      string(itemEntity.RarityCommon),
		IsActive:    true,
	})
	assert.NoError(t, err)

	// Test multiple reward items addition
	rewardItems := []itemEntity.RewardItem{
		{ItemID: sword.ID, Count: 1},
		{ItemID: gold.ID, Count: 100},
	}

	err = suite.itemService.AddMultipleToInventory(userID, rewardItems, "quest_completion")
	assert.NoError(t, err)

	// Verify inventory
	inventory, err := suite.itemService.GetUserInventory(userID)
	assert.NoError(t, err)
	assert.NotNil(t, inventory)
	assert.GreaterOrEqual(t, inventory.TotalItems, 2)

	// Check specific items in inventory
	foundSword := false
	foundGold := false
	for _, item := range inventory.Items {
		if item.ItemID == sword.ID {
			foundSword = true
			assert.GreaterOrEqual(t, item.Count, 1)
		}
		if item.ItemID == gold.ID {
			foundGold = true
			assert.GreaterOrEqual(t, item.Count, 100)
		}
	}
	assert.True(t, foundSword, "Sword not found in inventory")
	assert.True(t, foundGold, "Gold not found in inventory")
}

func (suite *IntegrationTestSuite) TestConcurrentUserCreation() {
	t := suite.T()

	const numUsers = 10
	results := make(chan error, numUsers)

	// Create multiple users concurrently
	for i := 0; i < numUsers; i++ {
		go func(index int) {
			request := userModule.CreateUserRequest{
				Name:     fmt.Sprintf("Concurrent User %d", index),
				Email:    fmt.Sprintf("concurrent%d@example.com", index),
				Age:      25 + index,
				Password: "password123",
			}

			_, err := suite.userService.CreateUser(request)
			results <- err
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < numUsers; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	// All operations should succeed
	assert.Empty(t, errors, "Concurrent user creation failed")

	// Verify all users were created
	users, err := suite.userService.ListUsers()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), numUsers)
}

func (suite *IntegrationTestSuite) TestErrorHandling() {
	t := suite.T()

	// Test user creation with duplicate email
	duplicateRequest := userModule.CreateUserRequest{
		Name:     "Duplicate User",
		Email:    suite.testUser.Email, // Same email as test user
		Age:      30,
		Password: "password123",
	}

	_, err := suite.userService.CreateUser(duplicateRequest)
	assert.Error(t, err)
	// Should be user exists error

	// Test getting non-existent user
	_, err = suite.userService.GetUser(999999)
	assert.Error(t, err)

	// Test getting non-existent item
	_, err = suite.itemService.GetItem(999999)
	assert.Error(t, err)

	// Test login with non-existent user
	_, err = suite.authService.Login("nonexistent@example.com", "password")
	assert.Error(t, err)
}

// HTTP API Integration Tests
func (suite *IntegrationTestSuite) TestHTTPAuthEndpoints() {
	t := suite.T()

	// Setup HTTP handlers (you'd typically do this in your main setup)
	authHandler := user.NewHandler(user.HandlerParam{
		AuthService: suite.authService,
		Validator:   validator.New(),
		Logger:      zap.NewNop(),
	})

	// Test login endpoint
	loginReq := user.LoginRequest{
		Email:    suite.testUser.Email,
		Password: suite.testUser.Password,
	}

	loginBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)

	err := authHandler.Login(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var loginResponse user.LoginResponse
	err = json.Unmarshal(rec.Body.Bytes(), &loginResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResponse.AccessToken)
	assert.NotEmpty(t, loginResponse.RefreshToken)

	// Test refresh endpoint
	refreshReq := user.RefreshTokenRequest{
		RefreshToken: loginResponse.RefreshToken,
	}

	refreshBody, _ := json.Marshal(refreshReq)
	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	c = suite.echo.NewContext(req, rec)

	err = authHandler.RefreshToken(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// Helper functions for setup
func setupInMemoryUserService(logger *zap.Logger) userModule.Service {
	// This would typically set up an in-memory repository
	// For simplicity, we're omitting the actual implementation
	// In real tests, you'd create actual repository instances
	return nil // Placeholder
}

func setupInMemoryAuthService(jwtService jwt.Service, userService userModule.Service, logger *zap.Logger) user.Service {
	// Similar setup for auth service
	return nil // Placeholder
}

func setupInMemoryItemService(logger *zap.Logger) itemModule.Service {
	// Similar setup for item service
	return nil // Placeholder
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// Performance tests
func BenchmarkUserCreation(b *testing.B) {
	logger := zap.NewNop()
	userService := setupInMemoryUserService(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := userModule.CreateUserRequest{
			Name:     fmt.Sprintf("Benchmark User %d", i),
			Email:    fmt.Sprintf("benchmark%d@example.com", i),
			Age:      25,
			Password: "password123",
		}

		_, err := userService.CreateUser(request)
		if err != nil {
			b.Fatalf("User creation failed: %v", err)
		}
	}
}

func BenchmarkPasswordHashing(b *testing.B) {
	password := "benchmarkPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := security.HashPassword(password, nil)
		if err != nil {
			b.Fatalf("Password hashing failed: %v", err)
		}
	}
}

func BenchmarkLogin(b *testing.B) {
	logger := zap.NewNop()
	jwtConfig := &jwt.Config{
		Secret:     "benchmark-secret",
		ExpiresIn:  time.Hour,
		TokenType:  "access",
	}
	jwtService := jwt.NewService(jwtConfig, logger)
	userService := setupInMemoryUserService(logger)
	authService := setupInMemoryAuthService(jwtService, userService, logger)

	// Create test user
	testUser := userModule.CreateUserRequest{
		Name:     "Benchmark User",
		Email:    "benchmark@example.com",
		Age:      25,
		Password: "password123",
	}
	userService.CreateUser(testUser)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := authService.Login(testUser.Email, testUser.Password)
		if err != nil {
			b.Fatalf("Login failed: %v", err)
		}
	}
}