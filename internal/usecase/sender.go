package usecase

import "github.com/go-playground/validator/v10"

type Subscribe struct {
	Email string `json:"email" validate:"required,email"`
}

func (s Subscribe) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}
