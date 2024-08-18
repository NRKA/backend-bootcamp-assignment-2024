package auth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/auth/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"github.com/jackc/pgx/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
)

type Repo struct {
	db *postgres.Database
}

func NewAuthRepo(db *postgres.Database) *Repo {
	return &Repo{
		db: db,
	}
}

func (repo *Repo) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var login usecase.LoginRequest
	err = json.Unmarshal(body, &login)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	response, err := repo.login(context.Background(), login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to login: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (repo *Repo) Register(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var login usecase.CreateUserRequest
	err = json.Unmarshal(body, &login)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
	}

	response, err := repo.register(context.Background(), login)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (repo *Repo) register(ctx context.Context, login usecase.CreateUserRequest) (usecase.CreateUserResponse, error) {
	var response usecase.CreateUserResponse
	token := gonanoid.Must()
	hash, err := HashPassword(login.Password)
	if err != nil {
		return response, err
	}

	query := `INSERT INTO user (email, password_hash, user_type, token) VALUES ($1, $2, $3, $4) RETURNING id`
	err = repo.db.ExecQueryRow(ctx, query, login.Email, hash, login.UserType, token).Scan(&response.UserId)

	if err != nil {
		return response, err
	}

	return response, nil
}

func (repo *Repo) login(ctx context.Context, login usecase.LoginRequest) (usecase.LoginResponse, error) {
	var response struct {
		Token        string `db:"token"`
		PasswordHash string `db:"password_hash"`
	}

	query := `SELECT token, password_hash FROM user WHERE id = $1`
	err := repo.db.Get(ctx, &response, query, login.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usecase.LoginResponse{}, ErrUserNotFound
		}
		return usecase.LoginResponse{}, err
	}

	if !CheckPassword(response.PasswordHash, login.Password) {
		return usecase.LoginResponse{}, errors.New("invalid password")
	}

	return usecase.LoginResponse{Token: response.Token}, nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
