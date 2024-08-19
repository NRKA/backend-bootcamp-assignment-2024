package usecase

import "github.com/go-playground/validator/v10"

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	UserType string `json:"user_type" validate:"required,oneof=client moderator"`
}

type CreateUserResponse struct {
	UserId int `json:"id"`
}

type LoginRequest struct {
	ID       int    `json:"id" validate:"required,gt=0"`
	Password string `json:"password" validate:"required"`
}

func (r CreateUserRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

func (r LoginRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type LoginResponse struct {
	Token string `json:"token"`
}
