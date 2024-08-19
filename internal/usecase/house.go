package usecase

import (
	"github.com/go-playground/validator/v10"
	"time"
)

type HouseCreateRequest struct {
	Address   string `json:"address" validate:"required"`
	Year      int    `json:"year" validate:"required,gte=0,lte=2100"`
	Developer string `json:"developer,omitempty" validate:"max=255"`
}

type House struct {
	ID        int       `json:"id"`
	Address   string    `json:"address"`
	Year      int       `json:"year"`
	Developer string    `json:"developer,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HouseFlats struct {
	Flat []FlatResponse
}

func (r HouseCreateRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
