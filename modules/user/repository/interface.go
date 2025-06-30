package repository

import (
	"errors"
	"fxserver/modules/user/entity"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserRepository interface {
	Create(user *entity.User) error
	GetByID(id int) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	Update(user *entity.User) error
	Delete(id int) error
	List() ([]*entity.User, error)
}