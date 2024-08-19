package usecase

import "github.com/go-playground/validator/v10"

type FlatCreateRequest struct {
	Number  int `json:"number" validate:"required,gt=0"`
	HouseID int `json:"house_id" validate:"required,gt=0"`
	Price   int `json:"price" validate:"required,gt=0"`
	Rooms   int `json:"rooms" validate:"required,gt=0"`
}

type FlatResponse struct {
	ID      int    `json:"id"`
	Number  int    `json:"number"`
	HouseID int    `json:"house_id"`
	Price   int    `json:"price"`
	Rooms   int    `json:"rooms"`
	Status  string `json:"status"`
}

type FlatUpdateRequest struct {
	ID     int    `json:"id" validate:"required,gt=0"`
	Status string `json:"status" validate:"required,oneof=on_moderate approved declined"`
}

func (r FlatCreateRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

func (r FlatUpdateRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
