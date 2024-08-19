package house

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

type houseRepoSuite struct {
	suite.Suite
	db   *postgres.Database
	repo *Repo
}

func TestHouseRepoSuite(t *testing.T) {
	suite.Run(t, new(houseRepoSuite))
}

func (suite *houseRepoSuite) SetupSuite() {
	db, err := postgres.NewDB(context.Background(), suite.fromEnv())
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewRepo(suite.db)
}

func (suite *houseRepoSuite) TearDownSuite() {
	suite.db.GetPool().Close()
}

func (suite *houseRepoSuite) SetupTest() {
	suite.clearTestDB(suite.db)
}

func (suite *houseRepoSuite) TestCreateSuccessful() {
	createRequest := usecase.HouseCreateRequest{
		Address: "someAddress",
		Year:    2000,
	}

	house, err := suite.repo.Create(context.Background(), createRequest)
	suite.Require().NoError(err)
	suite.Require().EqualValues(createRequest.Address, house.Address)
	suite.Require().EqualValues(createRequest.Year, house.Year)
}

func (suite *houseRepoSuite) TestClientFlatsSuccessful() {
	ctx := context.Background()
	createRequest := usecase.HouseCreateRequest{
		Address: "someAddress2",
		Year:    2000,
	}

	house, err := suite.repo.Create(context.Background(), createRequest)
	suite.Require().NoError(err)

	suite.insertTestFlat(ctx, house.ID, 1, 100000, 3, "approved")
	suite.insertTestFlat(ctx, house.ID, 2, 150000, 2, "created")

	flats, err := suite.repo.ClientFlats(ctx, house.ID)
	suite.Require().NoError(err)
	suite.Require().Len(flats, 1)
	suite.Require().Equal(1, flats[0].Number)
	suite.Require().Equal("approved", flats[0].Status)

}

func (suite *houseRepoSuite) TestModeratorFlatsSuccessful() {
	ctx := context.Background()
	createRequest := usecase.HouseCreateRequest{
		Address: "someAddress3",
		Year:    2000,
	}

	house, err := suite.repo.Create(context.Background(), createRequest)
	suite.Require().NoError(err)

	suite.insertTestFlat(ctx, house.ID, 1, 100000, 3, "approved")
	suite.insertTestFlat(ctx, house.ID, 2, 150000, 2, "created")

	flats, err := suite.repo.ModeratorFlats(ctx, house.ID)
	suite.Require().NoError(err)
	suite.Require().Len(flats, 2)
}

func (suite *houseRepoSuite) insertTestFlat(ctx context.Context, houseID, number, price, rooms int, status string) int {
	var flatID int
	err := suite.db.ExecQueryRow(ctx, `
		INSERT INTO flat (number, house_id, price, rooms, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, number, houseID, price, rooms, status).Scan(&flatID)
	suite.Require().NoError(err)
	return flatID
}

func (suite *houseRepoSuite) clearTestDB(db *postgres.Database) {
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

func (suite *houseRepoSuite) fromEnv() postgres.DatabaseConfig {
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
