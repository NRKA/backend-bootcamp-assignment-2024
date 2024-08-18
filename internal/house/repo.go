package house

import (
	"context"
	"encoding/json"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/house/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"io"
	"net/http"
)

type Repo struct {
	db *postgres.Database
}

func NewRepo(db *postgres.Database) *Repo {
	return &Repo{
		db: db,
	}
}

func (repo *Repo) Create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var houseRequest usecase.HouseCreateRequest
	err = json.Unmarshal(body, &houseRequest)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
	}

	response, err := repo.create(context.Background(), houseRequest)
	if err != nil {
		http.Error(w, "Failed to create house", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (repo *Repo) create(ctx context.Context, house usecase.HouseCreateRequest) (usecase.House, error) {
	var response usecase.House

	query := `
		INSERT INTO houses (address, year, developer) 
		VALUES ($1, $2, $3) 
		RETURNING *
	`

	err := repo.db.Get(ctx, &response, query, house.Address, house.Year, house.Developer)

	if err != nil {
		return response, err
	}

	return response, nil
}
