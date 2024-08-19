package house

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/auth"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/flat"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

type houseHandlerSuite struct {
	suite.Suite
	db      *postgres.Database
	auth    *auth.Handler
	flat    *flat.Handler
	handler *Handler
}

func TestHouseHandlerSuite(t *testing.T) {
	suite.Run(t, new(houseHandlerSuite))
}

func (suite *houseHandlerSuite) SetupSuite() {
	db, err := postgres.NewDB(context.Background(), suite.fromEnv())
	suite.Require().NoError(err)

	suite.db = db
	suite.auth = auth.NewHandler(db)
	suite.flat = flat.NewHandler(db)
	suite.handler = NewHandler(db)
}

func (suite *houseHandlerSuite) TearDownSuite() {
	suite.db.GetPool().Close()
}

func (suite *houseHandlerSuite) SetupTest() {
	suite.clearTestDB(suite.db)
}

func (suite *houseHandlerSuite) TestCreateHouseFailDecoding() {
	invalidJSON := `{"address": "123 Test Street", "year": "invalid_year", "developer": "Test Developer"`

	req := httptest.NewRequest(http.MethodPost, "/house/create", bytes.NewReader([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.clientToken())

	w := httptest.NewRecorder()
	suite.handler.Create(w, req)
	suite.Require().EqualValues(http.StatusBadRequest, w.Code)
	suite.Require().Contains(w.Body.String(), "Invalid request payload")
}

func (suite *houseHandlerSuite) TestCreateHouseFailValidation() {
	invalidRequest := `{"address": "Лесная улица, 7, Москва, 125196", "year": 2200, "developer": "Мэрия города"}`

	req := httptest.NewRequest(http.MethodPost, "/house/create", bytes.NewReader([]byte(invalidRequest)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.clientToken())

	w := httptest.NewRecorder()
	suite.handler.Create(w, req)

	suite.Require().EqualValues(http.StatusBadRequest, w.Code)

	body := w.Body.String()
	suite.Require().Contains(body, "Key: 'HouseCreateRequest.Year' Error:Field validation for 'Year' failed on the 'lte' tag", "Response body should contain the validation error message for the 'year' field")
}

func (suite *houseHandlerSuite) TestClientCreateHouseSuccess() {
	validRequest := `{
		"address": "Лесная улица, 7, Москва, 125196",
		"year": 2000,
		"developer": "Мэрия города"
	}`

	req := httptest.NewRequest(http.MethodPost, "/house/create", bytes.NewReader([]byte(validRequest)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.clientToken())

	w := httptest.NewRecorder()
	suite.handler.Create(w, req)

	suite.Require().EqualValues(http.StatusOK, w.Code)

	var response usecase.House
	err := json.NewDecoder(w.Body).Decode(&response)
	suite.Require().NoError(err, "Failed to decode response body")

	suite.Require().NotZero(response.ID, "Expected ID to be non-zero")
	suite.Require().EqualValues("Лесная улица, 7, Москва, 125196", response.Address, "Address does not match")
	suite.Require().EqualValues(2000, response.Year, "Year does not match")
	suite.Require().EqualValues("Мэрия города", response.Developer, "Developer does not match")
}

func (suite *houseHandlerSuite) TestModeratorCreateHouseSuccess() {
	validRequest := `{
		"address": "Садовая улица, 15, Москва, 123456",
		"year": 2010,
		"developer": "Строительный трест"
	}`

	req := httptest.NewRequest(http.MethodPost, "/house/create", bytes.NewReader([]byte(validRequest)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())
	w := httptest.NewRecorder()
	suite.handler.Create(w, req)

	suite.Require().EqualValues(http.StatusOK, w.Code)

	var response usecase.House
	err := json.NewDecoder(w.Body).Decode(&response)
	suite.Require().NoError(err, "Failed to decode response body")

	suite.Require().NotZero(response.ID, "Expected ID to be non-zero")
	suite.Require().EqualValues("Садовая улица, 15, Москва, 123456", response.Address, "Address does not match")
	suite.Require().EqualValues(2010, response.Year, "Year does not match")
	suite.Require().EqualValues("Строительный трест", response.Developer, "Developer does not match")
}

func (suite *houseHandlerSuite) TestClientFlats() {
	houseID := suite.createHouse()

	flats := []usecase.FlatCreateRequest{
		{Number: 101, HouseID: houseID, Price: 1500, Rooms: 3},
		{Number: 102, HouseID: houseID, Price: 2000, Rooms: 4},
	}

	var createdFlats []usecase.FlatResponse

	for _, flat := range flats {
		createRequestBody, _ := json.Marshal(flat)
		createReq := httptest.NewRequest(http.MethodPost, "/flat/create", bytes.NewReader(createRequestBody))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+suite.moderatorToken())
		createRecorder := httptest.NewRecorder()
		suite.flat.Create(createRecorder, createReq)

		suite.Require().EqualValues(http.StatusOK, createRecorder.Code)
		var createdFlat usecase.FlatResponse
		err := json.NewDecoder(createRecorder.Body).Decode(&createdFlat)
		suite.Require().NoError(err)
		createdFlats = append(createdFlats, createdFlat)
	}

	updateRequest := usecase.FlatUpdateRequest{
		ID:     createdFlats[0].ID,
		Status: "approved",
	}

	updateRequestBody, _ := json.Marshal(updateRequest)
	updateReq := httptest.NewRequest(http.MethodPut, "/flat/update", bytes.NewReader(updateRequestBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+suite.moderatorToken())
	updateRecorder := httptest.NewRecorder()
	suite.flat.Update(updateRecorder, updateReq)

	suite.Require().EqualValues(http.StatusOK, updateRecorder.Code)

	path := fmt.Sprintf("/house/")
	listReq := httptest.NewRequest(http.MethodGet, path, nil)
	listReq.SetPathValue("id", strconv.Itoa(houseID))
	listReq.Header.Set("Authorization", "Bearer "+suite.clientToken())
	listRecorder := httptest.NewRecorder()

	claims := &auth.Claims{
		UserID: 1,
		Role:   "client",
	}
	ctx := context.WithValue(listReq.Context(), "claims", claims)
	listReq = listReq.WithContext(ctx)

	suite.handler.Flats(listRecorder, listReq)

	suite.Require().EqualValues(http.StatusOK, listRecorder.Code)

	var flatsList usecase.HouseFlats
	err := json.NewDecoder(listRecorder.Body).Decode(&flatsList)
	suite.Require().NoError(err)

	suite.Require().Len(flatsList.Flat, 1)
	suite.Require().EqualValues(createdFlats[0].ID, flatsList.Flat[0].ID)
}

func (suite *houseHandlerSuite) TestModeratorFlats() {
	houseID := suite.createHouse()

	flats := []usecase.FlatCreateRequest{
		{Number: 101, HouseID: houseID, Price: 1500, Rooms: 3},
		{Number: 102, HouseID: houseID, Price: 2000, Rooms: 4},
	}

	var createdFlats []usecase.FlatResponse

	for _, flat := range flats {
		createRequestBody, _ := json.Marshal(flat)
		createReq := httptest.NewRequest(http.MethodPost, "/flat/create", bytes.NewReader(createRequestBody))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+suite.moderatorToken())
		createRecorder := httptest.NewRecorder()
		suite.flat.Create(createRecorder, createReq)

		suite.Require().EqualValues(http.StatusOK, createRecorder.Code)
		var createdFlat usecase.FlatResponse
		err := json.NewDecoder(createRecorder.Body).Decode(&createdFlat)
		suite.Require().NoError(err)
		createdFlats = append(createdFlats, createdFlat)
	}

	updateRequest := usecase.FlatUpdateRequest{
		ID:     createdFlats[0].ID,
		Status: "approved",
	}

	updateRequestBody, _ := json.Marshal(updateRequest)
	updateReq := httptest.NewRequest(http.MethodPut, "/flat/update", bytes.NewReader(updateRequestBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+suite.moderatorToken())
	updateRecorder := httptest.NewRecorder()
	suite.flat.Update(updateRecorder, updateReq)
	suite.Require().EqualValues(http.StatusOK, updateRecorder.Code)

	path := fmt.Sprintf("/house/")
	listReq := httptest.NewRequest(http.MethodGet, path, nil)
	listReq.SetPathValue("id", strconv.Itoa(houseID))
	listReq.Header.Set("Authorization", "Bearer "+suite.clientToken())
	listRecorder := httptest.NewRecorder()

	claims := &auth.Claims{
		UserID: 1,
		Role:   "moderator",
	}
	ctx := context.WithValue(listReq.Context(), "claims", claims)
	listReq = listReq.WithContext(ctx)

	suite.handler.Flats(listRecorder, listReq)
	suite.Require().Equal(http.StatusOK, listRecorder.Code)

	var flatsList usecase.HouseFlats
	err := json.NewDecoder(listRecorder.Body).Decode(&flatsList)
	suite.Require().NoError(err)

	suite.Require().Len(flatsList.Flat, 2)
}

func (suite *houseHandlerSuite) createHouse() int {
	validRequest := `{
		"address": "Садовая улица, 15, Москва, 123456",
		"year": 2010,
		"developer": "Строительный трест"
	}`

	req := httptest.NewRequest(http.MethodPost, "/house/create", bytes.NewReader([]byte(validRequest)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.moderatorToken())
	recorder := httptest.NewRecorder()
	suite.handler.Create(recorder, req)

	suite.Require().EqualValues(http.StatusOK, recorder.Code)

	var response usecase.House
	err := json.NewDecoder(recorder.Body).Decode(&response)
	suite.Require().NoError(err, "Failed to decode response body")

	return response.ID
}

func (suite *houseHandlerSuite) clientToken() string {
	req := httptest.NewRequest(http.MethodGet, "/dummyLogin?user_type=client", nil)
	w := httptest.NewRecorder()

	suite.auth.DummyLogin(w, req)

	resp := w.Result()
	suite.Require().Equal(http.StatusOK, resp.StatusCode)

	var loginResponse usecase.LoginResponse
	err := json.NewDecoder(resp.Body).Decode(&loginResponse)
	suite.Require().NoError(err)
	return loginResponse.Token
}

func (suite *houseHandlerSuite) moderatorToken() string {
	req := httptest.NewRequest(http.MethodGet, "/dummyLogin?user_type=moderator", nil)
	w := httptest.NewRecorder()

	suite.auth.DummyLogin(w, req)

	resp := w.Result()
	suite.Equal(http.StatusOK, resp.StatusCode)

	var loginResponse usecase.LoginResponse
	err := json.NewDecoder(resp.Body).Decode(&loginResponse)
	suite.Require().NoError(err)
	return loginResponse.Token
}

func (suite *houseHandlerSuite) clearTestDB(db *postgres.Database) {
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

func (suite *houseHandlerSuite) fromEnv() postgres.DatabaseConfig {
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
