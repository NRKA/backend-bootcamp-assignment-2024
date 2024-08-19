package sender

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"net/http"
	"strconv"
)

type Subscriber interface {
	Subscribe(ctx context.Context, houseID int, email usecase.Subscribe) error
}

type Handler struct {
	repo Subscriber
}

func NewHandler(db *postgres.Database) *Handler {
	return &Handler{NewRepo(db)}
}

func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be integer", http.StatusBadRequest)
		return
	}

	var email usecase.Subscribe
	if err = json.NewDecoder(r.Body).Decode(&email); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err = email.Validate(); err != nil {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	err = h.repo.Subscribe(r.Context(), id, email)
	if err != nil {
		if errors.Is(err, ErrHouseNotFound) {
			http.Error(w, "house with the given ID does not exist", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to subscribe", http.StatusInternalServerError)
		return
	}
}
