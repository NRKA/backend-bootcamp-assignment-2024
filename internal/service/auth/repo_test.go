package auth

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"testing"

	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"github.com/stretchr/testify/suite"
)

type authRepoSuite struct {
	suite.Suite
	repo *Repo
	db   *postgres.Database
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(authRepoSuite))
}

func (suite *authRepoSuite) SetupSuite() {
	db, err := postgres.NewDB(context.Background(), suite.fromEnv())
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewRepo(suite.db)
}

func (suite *authRepoSuite) TearDownSuite() {
	suite.db.GetPool().Close()
}

func (suite *authRepoSuite) SetupTest() {
	suite.clearTestDB(suite.db)
}

func (suite *authRepoSuite) TestRegisterSuccess() {
	ctx := context.Background()

	req := usecase.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		UserType: "client",
	}

	resp, err := suite.repo.Register(ctx, req)
	suite.Require().NoError(err)
	suite.Require().NotZero(resp.UserId)
}

func (suite *authRepoSuite) TestRegisterDuplicateEmail() {
	ctx := context.Background()

	req := usecase.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		UserType: "client",
	}

	_, err := suite.repo.Register(ctx, req)
	suite.Require().NoError(err)

	_, err = suite.repo.Register(ctx, req)
	suite.Require().Error(err)
}

func (suite *authRepoSuite) TestLoginSuccess() {
	ctx := context.Background()

	req := usecase.CreateUserRequest{
		Email:    "testlogin@example.com",
		Password: "password123",
		UserType: "client",
	}

	response, err := suite.repo.Register(ctx, req)
	suite.Require().NoError(err)

	loginReq := usecase.LoginRequest{
		ID:       response.UserId,
		Password: "password123",
	}

	resp, err := suite.repo.Login(ctx, loginReq)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(resp.Token)
}

func (suite *authRepoSuite) TestLoginInvalidPassword() {
	ctx := context.Background()

	req := usecase.CreateUserRequest{
		Email:    "testloginfail@example.com",
		Password: "password123",
		UserType: "client",
	}

	response, err := suite.repo.Register(ctx, req)
	suite.Require().NoError(err)

	loginReq := usecase.LoginRequest{
		ID:       response.UserId,
		Password: "wrongpassword",
	}

	_, err = suite.repo.Login(ctx, loginReq)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrInvalidPassword)
}

func (suite *authRepoSuite) clearTestDB(db *postgres.Database) {
	ctx := context.Background()
	tables := []string{`"user"`}
	_, err := db.Exec(ctx, "SET session_replication_role = 'replica'")
	suite.Require().NoError(err)

	for _, table := range tables {
		_, err := db.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
		suite.Require().NoError(err)
	}

	_, err = db.Exec(ctx, "SET session_replication_role = 'origin'")
	suite.Require().NoError(err)

	fmt.Println("Test database cleared successfully")
}

func (suite *authRepoSuite) fromEnv() postgres.DatabaseConfig {
	err := godotenv.Load("../../../.env")
	suite.Require().NoError(err)

	return postgres.DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
	}
}
