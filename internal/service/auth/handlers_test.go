package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type authHandlerSuite struct {
	suite.Suite
	db      *postgres.Database
	handler *Handler
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(authHandlerSuite))
}

func (suite *authHandlerSuite) SetupSuite() {
	db, err := postgres.NewDB(context.Background(), suite.fromEnv())
	suite.Require().NoError(err)

	suite.db = db
	suite.handler = NewHandler(db)
}

func (suite *authHandlerSuite) TearDownSuite() {
	suite.db.GetPool().Close()
}

func (suite *authHandlerSuite) SetupTest() {
	suite.clearTestDB(suite.db)
}

func (suite *authHandlerSuite) TestDummyLoginClientSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/dummyLogin?user_type=client", nil)
	w := httptest.NewRecorder()

	suite.handler.DummyLogin(w, req)

	resp := w.Result()
	suite.Require().EqualValues(http.StatusOK, resp.StatusCode)

	var loginResponse usecase.LoginResponse
	err := json.NewDecoder(resp.Body).Decode(&loginResponse)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(loginResponse.Token)
}

func (suite *authHandlerSuite) TestDummyLoginModeratorSuccess() {
	req := httptest.NewRequest(http.MethodGet, "/dummyLogin?user_type=moderator", nil)
	w := httptest.NewRecorder()

	suite.handler.DummyLogin(w, req)

	resp := w.Result()
	suite.Require().EqualValues(http.StatusOK, resp.StatusCode)

	var loginResponse usecase.LoginResponse
	err := json.NewDecoder(resp.Body).Decode(&loginResponse)
	suite.Require().NoError(err)
	suite.NotEmpty(loginResponse.Token)
	fmt.Println(loginResponse.Token)
}

func (suite *authHandlerSuite) TestDummyLoginInvalidRole() {
	req := httptest.NewRequest(http.MethodGet, "/dummyLogin?user_type=asdsadsad", nil)
	w := httptest.NewRecorder()

	suite.handler.DummyLogin(w, req)

	resp := w.Result()
	suite.Require().EqualValues(http.StatusBadRequest, resp.StatusCode)

	expectedErrorMessage := "Invalid role: Invalid request or missing user_type\n"
	suite.Require().EqualValues(expectedErrorMessage, w.Body.String())
}

func (suite *authHandlerSuite) TestRegisterFailDecoding() {
	invalidBody := []byte(`{"email": "test@example.com", "password": "short", "user_type": "client"`)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(invalidBody))
	w := httptest.NewRecorder()

	suite.handler.Register(w, req)

	resp := w.Result()
	suite.Require().EqualValues(http.StatusBadRequest, resp.StatusCode)

	expectedErrorMessage := "Invalid request payload\n"
	suite.Require().EqualValues(expectedErrorMessage, w.Body.String())
}

func (suite *authHandlerSuite) TestRegisterFailValidation() {
	invalidBody := []byte(`{
		"email": "invalid-email",
		"password": "short",
		"user_type": "invalid_type"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(invalidBody))
	w := httptest.NewRecorder()

	suite.handler.Register(w, req)

	resp := w.Result()
	suite.Require().EqualValues(http.StatusBadRequest, resp.StatusCode)

	expectedErrorMessage := "Key: 'CreateUserRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag\nKey: 'CreateUserRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag\nKey: 'CreateUserRequest.UserType' Error:Field validation for 'UserType' failed on the 'oneof' tag"
	suite.Require().Contains(w.Body.String(), expectedErrorMessage)
}

func (suite *authHandlerSuite) TestRegisterSuccess() {
	validBody := usecase.CreateUserRequest{
		Email:    "user@example.com",
		Password: "validpassword",
		UserType: "client",
	}
	validBodyBytes, err := json.Marshal(validBody)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(validBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.handler.Register(w, req)

	suite.Require().EqualValues(http.StatusOK, w.Code)

	var response usecase.CreateUserResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	suite.Greater(response.UserId, 0)
}

func (suite *authHandlerSuite) TestLoginFailDecoding() {
	invalidBody := `{ "id": "not_a_number", "password": "somepassword" }`
	invalidBodyBytes := []byte(invalidBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(invalidBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.handler.Login(w, req)

	suite.Require().EqualValues(http.StatusBadRequest, w.Code)

	expectedErrorMessage := "Invalid request payload"
	suite.Require().Contains(w.Body.String(), expectedErrorMessage)
}

func (suite *authHandlerSuite) TestLoginFailValidation() {
	invalidPayload := `{"id": 0, "password": "validpassword"}`
	expectedError := "Key: 'LoginRequest.ID' Error:Field validation for 'ID' failed on the 'required' tag\n"

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(invalidPayload)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.handler.Login(w, req)

	suite.Require().EqualValues(http.StatusBadRequest, w.Code)

	body := w.Body.String()
	suite.Require().Contains(body, expectedError)
}

func (suite *authHandlerSuite) TestLoginUserNotFound() {
	loginRequest := usecase.LoginRequest{
		ID:       999999999,
		Password: "somepassword",
	}

	payload, err := json.Marshal(loginRequest)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.handler.Login(w, req)

	suite.Require().EqualValues(http.StatusNotFound, w.Code)

	suite.Require().EqualValues("User not found\n", w.Body.String())
}

func (suite *authHandlerSuite) TestLoginInvalidPassword() {
	validBody := usecase.CreateUserRequest{
		Email:    "user@example.com",
		Password: "validpassword",
		UserType: "client",
	}
	validBodyBytes, err := json.Marshal(validBody)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(validBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.handler.Register(w, req)

	suite.Require().EqualValues(http.StatusOK, w.Code)

	var response usecase.CreateUserResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	loginRequest := usecase.LoginRequest{
		ID:       response.UserId,
		Password: "wrongpassword",
	}

	payload, err := json.Marshal(loginRequest)
	suite.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	suite.handler.Login(w, req)

	suite.Require().EqualValues(http.StatusUnauthorized, w.Code)
	suite.Require().EqualValues("Invalid password\n", w.Body.String())
}

func (suite *authHandlerSuite) clearTestDB(db *postgres.Database) {
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

func (suite *authHandlerSuite) fromEnv() postgres.DatabaseConfig {
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
