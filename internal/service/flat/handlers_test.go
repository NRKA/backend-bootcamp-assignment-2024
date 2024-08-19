package flat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/auth"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/house"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type flatHandlerSuite struct {
	suite.Suite
	db      *postgres.Database
	auth    *auth.Handler
	house   *house.Handler
	handler *Handler
}

func TestFlatHandlerSuite(t *testing.T) {
	suite.Run(t, new(flatHandlerSuite))
}

func (suite *flatHandlerSuite) SetupSuite() {
	db, err := postgres.NewDB(context.Background(), suite.fromEnv())
	suite.Require().NoError(err)

	suite.db = db
	suite.auth = auth.NewHandler(db)
	suite.house = house.NewHandler(db)
	suite.handler = NewHandler(db)
}

func (suite *flatHandlerSuite) SetupTest() {
	suite.clearTestDB(suite.db)
}

func (suite *flatHandlerSuite) TestCreateFailDecoding() {
	invalidRequestBody := `{"number": 101, "house_id": 1, "price": 1000, "rooms": "three"}`

	req := httptest.NewRequest(http.MethodPost, "/flat/create", bytes.NewReader([]byte(invalidRequestBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())

	w := httptest.NewRecorder()
	suite.handler.Create(w, req)
	suite.Require().EqualValues(http.StatusBadRequest, w.Code, "Expected status code 400 Bad Request")

	expectedErrorMessage := "Invalid request payload"
	suite.Require().Contains(w.Body.String(), expectedErrorMessage)
}

func (suite *flatHandlerSuite) TestCreateFailValidation() {
	invalidRequestBody := `{"number": 101, "house_id": 1, "price": -1000, "rooms": 3}`

	req := httptest.NewRequest(http.MethodPost, "/flat/create", bytes.NewReader([]byte(invalidRequestBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())

	w := httptest.NewRecorder()
	suite.handler.Create(w, req)

	suite.Require().EqualValues(http.StatusBadRequest, w.Code, "Expected status code 400 Bad Request")

	expectedErrorSubstring := "Field validation for 'Price' failed on the 'gt' tag"
	suite.Require().Contains(w.Body.String(), expectedErrorSubstring)
}

func (suite *flatHandlerSuite) TestCreateFlatHouseNotFound() {
	invalidRequestBody := `{"number": 101, "house_id": 99999, "price": 1000, "rooms": 3}`

	req := httptest.NewRequest(http.MethodPost, "/flat/create", bytes.NewReader([]byte(invalidRequestBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())

	w := httptest.NewRecorder()
	suite.handler.Create(w, req)

	suite.Require().EqualValues(http.StatusNotFound, w.Code, "Expected status code 404 Not Found")

	expectedErrorMessage := "House with this ID does not exist"
	suite.Contains(w.Body.String(), expectedErrorMessage)
}

func (suite *flatHandlerSuite) TestCreateFailDuplicateFlat() {
	houseRequest := `{"address": "Лесная улица, 7, Москва, 125196", "year": 2000, "developer": "Мэрия города"}`
	houseReq := httptest.NewRequest(http.MethodPost, "/house/create", bytes.NewReader([]byte(houseRequest)))
	houseReq.Header.Set("Content-Type", "application/json")
	houseReq.Header.Set("Authorization", "Bearer "+suite.moderatorToken())

	houseRecorder := httptest.NewRecorder()
	suite.house.Create(houseRecorder, houseReq)
	suite.Require().EqualValues(http.StatusOK, houseRecorder.Code, "Failed to create house")

	var houseResponse usecase.House
	err := json.NewDecoder(houseRecorder.Body).Decode(&houseResponse)
	suite.Require().NoError(err)

	validRequestBody := fmt.Sprintf(`{"number": 101, "house_id": %d, "price": 1000, "rooms": 3}`, houseResponse.ID)
	req := httptest.NewRequest(http.MethodPost, "/flat/create", bytes.NewReader([]byte(validRequestBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())

	recorder := httptest.NewRecorder()
	suite.handler.Create(recorder, req)

	suite.Require().EqualValues(http.StatusOK, recorder.Code, "Failed to create initial flat")

	duplicateRequestBody := fmt.Sprintf(`{"number": 101, "house_id": %d, "price": 1200, "rooms": 3}`, houseResponse.ID)
	duplicateReq := httptest.NewRequest(http.MethodPost, "/flat/create", bytes.NewReader([]byte(duplicateRequestBody)))
	duplicateReq.Header.Set("Content-Type", "application/json")
	duplicateReq.Header.Set("Authorization", "Bearer "+suite.moderatorToken())

	duplicateRecorder := httptest.NewRecorder()
	suite.handler.Create(duplicateRecorder, duplicateReq)

	suite.Require().EqualValues(http.StatusConflict, duplicateRecorder.Code)

	expectedErrorMessage := "House with this ID already exists"
	suite.Require().Contains(duplicateRecorder.Body.String(), expectedErrorMessage)
}

func (suite *flatHandlerSuite) TestCreateSuccess() {
	houseRequest := `{"address": "Лесная улица, 7, Москва, 125196", "year": 2000, "developer": "Мэрия города"}`
	houseReq := httptest.NewRequest(http.MethodPost, "/house/create", bytes.NewReader([]byte(houseRequest)))
	houseReq.Header.Set("Content-Type", "application/json")
	houseReq.Header.Set("Authorization", "Bearer "+suite.moderatorToken()) // Используем токен модератора

	houseRecorder := httptest.NewRecorder()
	suite.house.Create(houseRecorder, houseReq)
	suite.Require().EqualValues(http.StatusOK, houseRecorder.Code, "Failed to create house")

	var houseResponse usecase.House
	err := json.NewDecoder(houseRecorder.Body).Decode(&houseResponse)
	suite.Require().NoError(err)

	validRequestBody := fmt.Sprintf(`{"number": 101, "house_id": %d, "price": 1000, "rooms": 3}`, houseResponse.ID)
	req := httptest.NewRequest(http.MethodPost, "/flat/create", bytes.NewReader([]byte(validRequestBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())

	w := httptest.NewRecorder()
	suite.handler.Create(w, req)

	suite.Require().EqualValues(http.StatusOK, w.Code)

	var flatResponse usecase.FlatResponse
	err = json.NewDecoder(w.Body).Decode(&flatResponse)
	suite.Require().NoError(err)

	suite.Require().EqualValues(101, flatResponse.Number)
	suite.Require().EqualValues(houseResponse.ID, flatResponse.HouseID)
	suite.Require().EqualValues(1000, flatResponse.Price)
	suite.Require().EqualValues(3, flatResponse.Rooms)
	suite.Require().EqualValues("created", flatResponse.Status)
}

func (suite *flatHandlerSuite) TestUpdateFailDecoding() {
	invalidRequestBody := `{"id": "not_an_integer", "status": "approved"}`

	req := httptest.NewRequest(http.MethodPut, "/flat/update", bytes.NewReader([]byte(invalidRequestBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())

	w := httptest.NewRecorder()
	suite.handler.Update(w, req)

	suite.Require().EqualValues(http.StatusBadRequest, w.Code)

	expectedErrorMessage := "Invalid request payload"
	suite.Require().Contains(w.Body.String(), expectedErrorMessage)
}

func (suite *flatHandlerSuite) TestUpdateFailValidation() {
	invalidRequestBody := `{"id": 1, "status": "invalid_status"}`

	req := httptest.NewRequest(http.MethodPut, "/flat/update", bytes.NewReader([]byte(invalidRequestBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())

	w := httptest.NewRecorder()
	suite.handler.Update(w, req)

	suite.Require().EqualValues(http.StatusBadRequest, w.Code)

	errorMessage := w.Body.String()
	suite.Require().Contains(errorMessage, "Key: 'FlatUpdateRequest.Status' Error:Field validation for 'Status' failed on the 'oneof' tag")
}

func (suite *flatHandlerSuite) TestUpdateFlatNotFound() {
	invalidRequestBody := `{"id": 1, "status": "approved"}`

	req := httptest.NewRequest(http.MethodPut, "/flat/update", bytes.NewReader([]byte(invalidRequestBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())

	recorder := httptest.NewRecorder()
	suite.handler.Update(recorder, req)

	suite.Require().EqualValues(http.StatusNotFound, recorder.Code)

	expectedErrorMessage := "Flat does not exist"
	suite.Require().Contains(recorder.Body.String(), expectedErrorMessage)
}

func (suite *flatHandlerSuite) TestUpdateSuccess() {
	houseID := suite.createHouse()

	initialFlatRequest := usecase.FlatCreateRequest{
		Number:  101,
		HouseID: houseID,
		Price:   1500,
		Rooms:   3,
	}

	createRequestBody, _ := json.Marshal(initialFlatRequest)
	createReq := httptest.NewRequest(http.MethodPost, "/flat/create", bytes.NewReader(createRequestBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+suite.moderatorToken())
	createRecorder := httptest.NewRecorder()
	suite.handler.Create(createRecorder, createReq)

	suite.Require().EqualValues(http.StatusOK, createRecorder.Code)
	var createdFlat usecase.FlatResponse
	err := json.NewDecoder(createRecorder.Body).Decode(&createdFlat)
	suite.Require().NoError(err, "Failed to decode created flat response")

	updateRequest := usecase.FlatUpdateRequest{
		ID:     createdFlat.ID,
		Status: "approved",
	}

	updateRequestBody, _ := json.Marshal(updateRequest)
	updateReq := httptest.NewRequest(http.MethodPut, "/flat/update", bytes.NewReader(updateRequestBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+suite.moderatorToken())
	updateRecorder := httptest.NewRecorder()
	suite.handler.Update(updateRecorder, updateReq)

	suite.Require().EqualValues(http.StatusOK, updateRecorder.Code)

	var updatedFlat usecase.FlatResponse
	err = json.NewDecoder(updateRecorder.Body).Decode(&updatedFlat)
	suite.Require().NoError(err)

	suite.Require().EqualValues(createdFlat.ID, updatedFlat.ID)
	suite.Require().EqualValues("approved", updatedFlat.Status)
}

func (suite *flatHandlerSuite) createHouse() int {
	validRequest := `{
		"address": "Садовая улица, 15, Москва, 123456",
		"year": 2010,
		"developer": "Строительный трест"
	}`

	req := httptest.NewRequest(http.MethodPost, "/house/create", bytes.NewReader([]byte(validRequest)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())
	w := httptest.NewRecorder()
	suite.house.Create(w, req)

	suite.Require().EqualValues(http.StatusOK, w.Code)

	var response usecase.House
	err := json.NewDecoder(w.Body).Decode(&response)
	suite.Require().NoError(err, "Failed to decode response body")

	return response.ID
}

func (suite *flatHandlerSuite) moderatorToken() string {
	req := httptest.NewRequest(http.MethodGet, "/dummyLogin?user_type=moderator", nil)
	w := httptest.NewRecorder()

	suite.auth.DummyLogin(w, req)

	resp := w.Result()
	suite.Require().EqualValues(http.StatusOK, resp.StatusCode)

	var loginResponse usecase.LoginResponse
	err := json.NewDecoder(resp.Body).Decode(&loginResponse)
	suite.Require().NoError(err)
	return loginResponse.Token
}

func (suite *flatHandlerSuite) clearTestDB(db *postgres.Database) {
	ctx := context.Background()
	tables := []string{`"user"`, "flat", "house"}
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

func (suite *flatHandlerSuite) fromEnv() postgres.DatabaseConfig {
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
