package user

import "fxserver/modules/user/entity"

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"required,min=1,max=150"`
	Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	Name  string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
	Age   int    `json:"age,omitempty" validate:"omitempty,min=1,max=150"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type ListUsersResponse struct {
	Users []entity.UserResponse `json:"users"`
	Total int                   `json:"total"`
}