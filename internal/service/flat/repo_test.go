package flat

import (
	"context"
	"fmt"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type flatRepoSuite struct {
	suite.Suite
	db   *postgres.Database
	repo *Repo
}

func TestFlatRepoSuite(t *testing.T) {
	suite.Run(t, new(flatRepoSuite))
}

func (suite *flatRepoSuite) SetupSuite() {
	db, err := postgres.NewDB(context.Background(), suite.fromEnv())
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewRepo(suite.db)
}

func (suite *flatRepoSuite) TearDownSuite() {
	suite.db.GetPool().Close()
}

func (suite *flatRepoSuite) SetupTest() {
	suite.clearTestDB(suite.db)
}

func (suite *flatRepoSuite) TestCreateFlatSuccessful() {
	ctx := context.Background()
	houseID := suite.insertTestHouse(ctx, "123 Test St", 2022)
	createRequest := usecase.FlatCreateRequest{
		Number:  1,
		HouseID: houseID,
		Price:   100000,
		Rooms:   3,
	}

	flat, err := suite.repo.Create(ctx, createRequest)
	suite.Require().NoError(err)
	suite.Require().Equal(createRequest.Number, flat.Number)
	suite.Require().Equal(createRequest.HouseID, flat.HouseID)
	suite.Require().Equal(createRequest.Price, flat.Price)
	suite.Require().Equal(createRequest.Rooms, flat.Rooms)
	suite.Require().NotZero(flat.ID)
}

func (suite *flatRepoSuite) TestCreateFlatHouseNotFound() {
	ctx := context.Background()
	createRequest := usecase.FlatCreateRequest{
		Number:  1,
		HouseID: 99999999,
		Price:   100000,
		Rooms:   3,
	}

	flat, err := suite.repo.Create(ctx, createRequest)
	suite.Require().Error(err)
	suite.Require().Equal(ErrHouseNotFound, err)
	suite.Require().Zero(flat.ID)
}

func (suite *flatRepoSuite) TestCreateDuplicateFlat() {
	ctx := context.Background()
	houseID := suite.insertTestHouse(ctx, "123 Test St", 2022)
	createRequest := usecase.FlatCreateRequest{
		Number:  1,
		HouseID: houseID,
		Price:   100000,
		Rooms:   3,
	}

	_, err := suite.repo.Create(ctx, createRequest)
	suite.Require().NoError(err)

	_, err = suite.repo.Create(ctx, createRequest)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrDuplicateFlat)
}

func (suite *flatRepoSuite) TestUpdateSuccessful() {
	ctx := context.Background()
	houseID := suite.insertTestHouse(ctx, "123 Test Street", 2022)
	flat, err := suite.repo.Create(ctx, usecase.FlatCreateRequest{
		Number:  1123,
		HouseID: houseID,
		Price:   100000,
		Rooms:   3,
	})
	suite.Require().NoError(err)

	updateRequest := usecase.FlatUpdateRequest{
		ID:     flat.ID,
		Status: "approved",
	}

	flat, err = suite.repo.Update(ctx, updateRequest)
	suite.Require().NoError(err)
	suite.Require().EqualValues(updateRequest.Status, flat.Status)
}

func (suite *flatRepoSuite) TestUpdateFlatNotFound() {
	ctx := context.Background()

	updateRequest := usecase.FlatUpdateRequest{
		ID:     9999999,
		Status: "approved",
	}

	_, err := suite.repo.Update(ctx, updateRequest)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrFlatNotFound)
}

func (suite *flatRepoSuite) insertTestHouse(ctx context.Context, address string, year int) int {
	var houseID int
	err := suite.db.ExecQueryRow(ctx, `
		INSERT INTO house (address, year)
		VALUES ($1, $2)
		RETURNING id
	`, address, year).Scan(&houseID)
	suite.Require().NoError(err)
	return houseID
}

func (suite *flatRepoSuite) clearTestDB(db *postgres.Database) {
	ctx := context.Background()
	tables := []string{"house", "flat"} // Укажите все таблицы, которые нужно очистить
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

func (suite *flatRepoSuite) fromEnv() postgres.DatabaseConfig {
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
