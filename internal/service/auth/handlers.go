package auth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"net/http"
)

const (
	userType  = "user_type"
	client    = "client"
	moderator = "moderator"

	id = 1
)

type Authorizer interface {
	Register(ctx context.Context, login usecase.CreateUserRequest) (usecase.CreateUserResponse, error)
	Login(ctx context.Context, login usecase.LoginRequest) (usecase.LoginResponse, error)
}

type Handler struct {
	repo Authorizer
}

func NewHandler(db *postgres.Database) *Handler {
	return &Handler{repo: NewRepo(db)}
}

func (auth *Handler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get(userType)
	if role != client && role != moderator {
		http.Error(w, "Invalid role: Invalid request or missing user_type", http.StatusBadRequest)
		return
	}

	token, err := GenerateToken(id, role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(usecase.LoginResponse{Token: token}); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (auth *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := auth.repo.Register(r.Context(), req)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (auth *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req usecase.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := auth.repo.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrInvalidPassword) {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		http.Error(w, "Failed to login", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
