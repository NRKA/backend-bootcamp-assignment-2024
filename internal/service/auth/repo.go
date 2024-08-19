package auth

import (
	"context"
	"errors"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var _ Authorizer = (*Repo)(nil)

type Repo struct {
	db *postgres.Database
}

func NewRepo(db *postgres.Database) *Repo {
	return &Repo{
		db: db,
	}
}

func (repo *Repo) Register(ctx context.Context, login usecase.CreateUserRequest) (usecase.CreateUserResponse, error) {
	var response usecase.CreateUserResponse
	hash, err := hashPassword(login.Password)
	if err != nil {
		return response, err
	}

	query := `INSERT INTO "user" (email, password_hash, user_type) VALUES ($1, $2, $3) RETURNING id`
	err = repo.db.ExecQueryRow(ctx, query, login.Email, hash, login.UserType).Scan(&response.UserId)

	if err != nil {
		return response, err
	}

	return response, nil
}

func (repo *Repo) Login(ctx context.Context, login usecase.LoginRequest) (usecase.LoginResponse, error) {
	var response struct {
		UserType     string `db:"user_type"`
		PasswordHash string `db:"password_hash"`
	}

	query := `SELECT password_hash, user_type FROM "user" WHERE id = $1`
	err := repo.db.Get(ctx, &response, query, login.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usecase.LoginResponse{}, ErrUserNotFound
		}

		return usecase.LoginResponse{}, err
	}

	if !checkPassword(response.PasswordHash, login.Password) {
		return usecase.LoginResponse{}, ErrInvalidPassword
	}

	token, err := GenerateToken(login.ID, response.UserType)
	if err != nil {
		return usecase.LoginResponse{}, err
	}

	return usecase.LoginResponse{Token: token}, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
