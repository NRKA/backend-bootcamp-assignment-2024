package house

import (
	"context"
	"encoding/json"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/auth"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"net/http"
	"strconv"
)

type House interface {
	Create(ctx context.Context, house usecase.HouseCreateRequest) (usecase.House, error)
	ClientFlats(ctx context.Context, houseID int) ([]usecase.FlatResponse, error)
	ModeratorFlats(ctx context.Context, houseID int) ([]usecase.FlatResponse, error)
}

type Handler struct {
	repo House
}

func NewHandler(db *postgres.Database) *Handler {
	return &Handler{NewRepo(db)}
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req usecase.HouseCreateRequest
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
		http.Error(w, "Failed to create house", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to create house", http.StatusInternalServerError)
	}
}

func (h Handler) Flats(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be integer", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value("claims").(*auth.Claims)
	if !ok {
		http.Error(w, "Invalid or missing user claims", http.StatusUnauthorized)
		return
	}

	var response []usecase.FlatResponse
	role := claims.Role
	if role == moderator {
		response, err = h.repo.ModeratorFlats(r.Context(), id)
	} else {
		response, err = h.repo.ClientFlats(r.Context(), id)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(usecase.HouseFlats{Flat: response}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
