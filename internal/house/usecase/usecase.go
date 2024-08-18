package usecase

import "time"

type HouseCreateRequest struct {
	Address   string `json:"address"`
	Year      int    `json:"year"`
	Developer string `json:"developer,omitempty"`
}

type House struct {
	ID        int       `json:"id"`
	Address   string    `json:"address"`
	Year      int       `json:"year"`
	Developer string    `json:"developer,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
