package flat

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"net/http"
)

type Flat interface {
	Create(ctx context.Context, request usecase.FlatCreateRequest) (usecase.FlatResponse, error)
	Update(ctx context.Context, request usecase.FlatUpdateRequest) (usecase.FlatResponse, error)
}

type Handler struct {
	repo Flat
}

func NewHandler(db *postgres.Database) *Handler {
	return &Handler{repo: NewRepo(db)}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req usecase.FlatCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.repo.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrHouseNotFound) {
			http.Error(w, "House with this ID does not exist", http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrDuplicateFlat) {
			http.Error(w, "House with this ID already exists", http.StatusConflict)
			return
		}

		http.Error(w, "Failed to create flat", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to create flat", http.StatusInternalServerError)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req usecase.FlatUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.repo.Update(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrFlatNotFound) {
			http.Error(w, "Flat does not exist", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to update flat", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to update flat", http.StatusInternalServerError)
	}
}
