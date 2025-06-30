package user

import (
	"errors"
	"testing"

	"fxserver/modules/user/entity"
	"fxserver/modules/user/repository"
	"fxserver/pkg/security"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock repository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *entity.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id int) (*entity.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*entity.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *entity.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List() ([]*entity.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func setupUserService(mockRepo repository.UserRepository) Service {
	logger := zap.NewNop()
	return &service{
		repo:   mockRepo,
		logger: logger,
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateUserRequest
		setupMock   func(*MockUserRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name: "successful user creation",
			request: CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      30,
				Password: "password123",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("Create", mock.AnythingOfType("*entity.User")).Return(nil).Run(func(args mock.Arguments) {
					user := args.Get(0).(*entity.User)
					user.ID = 1 // Simulate ID assignment
				})
			},
			wantErr: false,
		},
		{
			name: "user already exists",
			request: CreateUserRequest{
				Name:     "Jane Doe",
				Email:    "existing@example.com",
				Age:      25,
				Password: "password123",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("Create", mock.AnythingOfType("*entity.User")).Return(repository.ErrUserExists)
			},
			wantErr:     true,
			wantErrType: repository.ErrUserExists,
		},
		{
			name: "repository error",
			request: CreateUserRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Age:      20,
				Password: "password123",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("Create", mock.AnythingOfType("*entity.User")).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			
			service := setupUserService(mockRepo)
			
			user, err := service.CreateUser(tt.request)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.request.Name, user.Name)
				assert.Equal(t, tt.request.Email, user.Email)
				assert.Equal(t, tt.request.Age, user.Age)
				// Password should be hashed, not plain text
				assert.NotEqual(t, tt.request.Password, user.Password)
				assert.NotEmpty(t, user.Password)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		setupMock   func(*MockUserRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name:   "successful user retrieval",
			userID: 1,
			setupMock: func(m *MockUserRepository) {
				user := &entity.User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
					Age:   30,
				}
				m.On("GetByID", 1).Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", 999).Return(nil, repository.ErrUserNotFound)
			},
			wantErr:     true,
			wantErrType: repository.ErrUserNotFound,
		},
		{
			name:   "repository error",
			userID: 1,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", 1).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			
			service := setupUserService(mockRepo)
			
			user, err := service.GetUser(tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	existingUser := &entity.User{
		ID:       1,
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Password: "hashedPassword123",
	}

	tests := []struct {
		name        string
		userID      int
		request     UpdateUserRequest
		setupMock   func(*MockUserRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name:   "successful user update",
			userID: 1,
			request: UpdateUserRequest{
				Name:  "John Updated",
				Email: "john.updated@example.com",
				Age:   31,
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", 1).Return(existingUser, nil)
				m.On("Update", mock.AnythingOfType("*entity.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "update with password",
			userID: 1,
			request: UpdateUserRequest{
				Name:     "John Updated",
				Password: "newPassword123",
			},
			setupMock: func(m *MockUserRepository) {
				// Create a copy of existing user to avoid modifying the original
				userCopy := *existingUser
				m.On("GetByID", 1).Return(&userCopy, nil)
				m.On("Update", mock.AnythingOfType("*entity.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			request: UpdateUserRequest{
				Name: "Updated Name",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", 999).Return(nil, repository.ErrUserNotFound)
			},
			wantErr:     true,
			wantErrType: repository.ErrUserNotFound,
		},
		{
			name:   "email already exists",
			userID: 1,
			request: UpdateUserRequest{
				Email: "existing@example.com",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", 1).Return(existingUser, nil)
				m.On("Update", mock.AnythingOfType("*entity.User")).Return(repository.ErrUserExists)
			},
			wantErr:     true,
			wantErrType: repository.ErrUserExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			
			service := setupUserService(mockRepo)
			
			user, err := service.UpdateUser(tt.userID, tt.request)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				
				if tt.request.Name != "" {
					assert.Equal(t, tt.request.Name, user.Name)
				}
				if tt.request.Email != "" {
					assert.Equal(t, tt.request.Email, user.Email)
				}
				if tt.request.Age != 0 {
					assert.Equal(t, tt.request.Age, user.Age)
				}
				if tt.request.Password != "" {
					// Password should be hashed, not plain text
					assert.NotEqual(t, tt.request.Password, user.Password)
					// New password hash should be different from old hash
					assert.NotEqual(t, existingUser.Password, user.Password)
					// Verify new password can be verified
					isValid, err := security.VerifyPassword(tt.request.Password, user.Password)
					assert.NoError(t, err)
					assert.True(t, isValid)
				}
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestVerifyUserPassword(t *testing.T) {
	// Create a test user with hashed password
	plainPassword := "testPassword123"
	hashedPassword, err := security.HashPassword(plainPassword, nil)
	assert.NoError(t, err)

	testUser := &entity.User{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Password: hashedPassword,
	}

	tests := []struct {
		name        string
		email       string
		password    string
		setupMock   func(*MockUserRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name:     "successful password verification",
			email:    "test@example.com",
			password: plainPassword,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByEmail", "test@example.com").Return(testUser, nil)
			},
			wantErr: false,
		},
		{
			name:     "wrong password",
			email:    "test@example.com",
			password: "wrongPassword",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByEmail", "test@example.com").Return(testUser, nil)
			},
			wantErr: true,
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByEmail", "nonexistent@example.com").Return(nil, repository.ErrUserNotFound)
			},
			wantErr: true,
		},
		{
			name:     "repository error",
			email:    "test@example.com",
			password: "password",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByEmail", "test@example.com").Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			
			service := setupUserService(mockRepo)
			
			user, err := service.VerifyUserPassword(tt.email, tt.password)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, testUser.ID, user.ID)
				assert.Equal(t, testUser.Email, user.Email)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		setupMock   func(*MockUserRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name:   "successful user deletion",
			userID: 1,
			setupMock: func(m *MockUserRepository) {
				m.On("Delete", 1).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func(m *MockUserRepository) {
				m.On("Delete", 999).Return(repository.ErrUserNotFound)
			},
			wantErr:     true,
			wantErrType: repository.ErrUserNotFound,
		},
		{
			name:   "repository error",
			userID: 1,
			setupMock: func(m *MockUserRepository) {
				m.On("Delete", 1).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			
			service := setupUserService(mockRepo)
			
			err := service.DeleteUser(tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
			} else {
				assert.NoError(t, err)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListUsers(t *testing.T) {
	testUsers := []*entity.User{
		{ID: 1, Name: "User 1", Email: "user1@example.com", Age: 25},
		{ID: 2, Name: "User 2", Email: "user2@example.com", Age: 30},
	}

	tests := []struct {
		name      string
		setupMock func(*MockUserRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name: "successful user list",
			setupMock: func(m *MockUserRepository) {
				m.On("List").Return(testUsers, nil)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "empty user list",
			setupMock: func(m *MockUserRepository) {
				m.On("List").Return([]*entity.User{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "repository error",
			setupMock: func(m *MockUserRepository) {
				m.On("List").Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			
			service := setupUserService(mockRepo)
			
			users, err := service.ListUsers()
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, users)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.Len(t, users, tt.wantCount)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

// Integration test with real password hashing
func TestCreateUserWithRealPasswordHashing(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(0).(*entity.User)
		user.ID = 1
	})

	service := setupUserService(mockRepo)

	request := CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Password: "testPassword123",
	}

	user, err := service.CreateUser(request)

	assert.NoError(t, err)
	assert.NotNil(t, user)

	// Verify password is hashed correctly
	isValid, err := security.VerifyPassword(request.Password, user.Password)
	assert.NoError(t, err)
	assert.True(t, isValid)

	// Verify wrong password fails
	isValid, err = security.VerifyPassword("wrongPassword", user.Password)
	assert.NoError(t, err)
	assert.False(t, isValid)

	mockRepo.AssertExpectations(t)
}