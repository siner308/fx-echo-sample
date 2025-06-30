package user

import (
	"errors"

	"fxserver/modules/user/entity"
	"fxserver/modules/user/repository"
	"fxserver/pkg/security"

	"go.uber.org/zap"
)

type Service interface {
	CreateUser(req CreateUserRequest) (*entity.User, error)
	GetUser(id int) (*entity.User, error)
	UpdateUser(id int, req UpdateUserRequest) (*entity.User, error)
	DeleteUser(id int) error
	ListUsers() ([]*entity.User, error)
	VerifyUserPassword(email, password string) (*entity.User, error)
}

type service struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

func NewService(repo repository.UserRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) CreateUser(req CreateUserRequest) (*entity.User, error) {
	// Hash the password before storing
	hashedPassword, err := security.HashPassword(req.Password, nil)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, errors.New("failed to process password")
	}

	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Age:      req.Age,
		Password: hashedPassword,
	}

	if err := s.repo.Create(user); err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			s.logger.Warn("Attempt to create user with existing email", zap.String("email", req.Email))
			return nil, err
		}
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	s.logger.Info("User created successfully", zap.Int("user_id", user.ID))
	return user, nil
}

func (s *service) GetUser(id int) (*entity.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Warn("User not found", zap.Int("user_id", id))
			return nil, err
		}
		s.logger.Error("Failed to get user", zap.Int("user_id", id), zap.Error(err))
		return nil, err
	}

	return user, nil
}

func (s *service) UpdateUser(id int, req UpdateUserRequest) (*entity.User, error) {
	existingUser, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Warn("User not found for update", zap.Int("user_id", id))
			return nil, err
		}
		s.logger.Error("Failed to get user for update", zap.Int("user_id", id), zap.Error(err))
		return nil, err
	}

	// Update only provided fields
	if req.Name != "" {
		existingUser.Name = req.Name
	}
	if req.Email != "" {
		existingUser.Email = req.Email
	}
	if req.Age != 0 {
		existingUser.Age = req.Age
	}
	if req.Password != "" {
		// Hash the new password
		hashedPassword, err := security.HashPassword(req.Password, nil)
		if err != nil {
			s.logger.Error("Failed to hash password during update", zap.Error(err))
			return nil, errors.New("failed to process password")
		}
		existingUser.Password = hashedPassword
	}

	if err := s.repo.Update(existingUser); err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			s.logger.Warn("Attempt to update user with existing email", zap.String("email", req.Email))
			return nil, err
		}
		s.logger.Error("Failed to update user", zap.Int("user_id", id), zap.Error(err))
		return nil, err
	}

	s.logger.Info("User updated successfully", zap.Int("user_id", id))
	return existingUser, nil
}

func (s *service) DeleteUser(id int) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Warn("User not found for deletion", zap.Int("user_id", id))
			return err
		}
		s.logger.Error("Failed to delete user", zap.Int("user_id", id), zap.Error(err))
		return err
	}

	s.logger.Info("User deleted successfully", zap.Int("user_id", id))
	return nil
}

func (s *service) ListUsers() ([]*entity.User, error) {
	users, err := s.repo.List()
	if err != nil {
		s.logger.Error("Failed to list users", zap.Error(err))
		return nil, err
	}

	return users, nil
}

// VerifyUserPassword verifies user credentials by email and password
func (s *service) VerifyUserPassword(email, password string) (*entity.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Warn("Login attempt with non-existent email", zap.String("email", email))
			return nil, errors.New("invalid credentials")
		}
		s.logger.Error("Failed to get user by email", zap.String("email", email), zap.Error(err))
		return nil, errors.New("authentication failed")
	}

	// Verify password
	isValid, err := security.VerifyPassword(password, user.Password)
	if err != nil {
		s.logger.Error("Failed to verify password", zap.Error(err))
		return nil, errors.New("authentication failed")
	}

	if !isValid {
		s.logger.Warn("Invalid password attempt", zap.String("email", email))
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
