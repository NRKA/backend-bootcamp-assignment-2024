package flat

import (
	"context"
	"errors"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ Flat = (*Repo)(nil)

type Repo struct {
	db *postgres.Database
}

func NewRepo(db *postgres.Database) *Repo {
	return &Repo{
		db: db,
	}
}

func (repo *Repo) Create(ctx context.Context, request usecase.FlatCreateRequest) (usecase.FlatResponse, error) {
	tx, err := repo.db.BeginTx(ctx)
	if err != nil {
		return usecase.FlatResponse{}, err
	}
	defer repo.db.RollbackTx(ctx, tx)

	var houseExists bool
	row := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM house WHERE id=$1)", request.HouseID)
	err = row.Scan(&houseExists)
	if err != nil {
		return usecase.FlatResponse{}, err
	}

	if !houseExists {
		return usecase.FlatResponse{}, ErrHouseNotFound
	}

	query := `
		INSERT INTO flat (number, house_id, price, rooms)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`
	var response usecase.FlatResponse
	row = tx.QueryRow(ctx, query, request.Number, request.HouseID, request.Price, request.Rooms)
	err = row.Scan(&response.ID, &response.Number, &response.HouseID, &response.Price, &response.Rooms, &response.Status)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return usecase.FlatResponse{}, ErrDuplicateFlat
		}
		return usecase.FlatResponse{}, err
	}

	updateHouseQuery := `
		UPDATE house
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, updateHouseQuery, request.HouseID)
	if err != nil {
		return usecase.FlatResponse{}, err
	}

	err = repo.db.CommitTx(ctx, tx)
	if err != nil {
		return usecase.FlatResponse{}, err
	}

	return response, nil
}

func (repo *Repo) Update(ctx context.Context, request usecase.FlatUpdateRequest) (usecase.FlatResponse, error) {
	tx, err := repo.db.BeginTx(ctx)
	if err != nil {
		return usecase.FlatResponse{}, err
	}
	defer repo.db.RollbackTx(ctx, tx)

	query := `
		UPDATE flat
		SET status = $1
		WHERE id = $2
		RETURNING *
	`

	var response usecase.FlatResponse
	row := tx.QueryRow(ctx, query, request.Status, request.ID)
	err = row.Scan(&response.ID, &response.Number, &response.HouseID, &response.Price, &response.Rooms, &response.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usecase.FlatResponse{}, ErrFlatNotFound
		}
		return usecase.FlatResponse{}, err
	}

	updateHouseQuery := `
		UPDATE house
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = (SELECT house_id FROM flat WHERE id = $1)
	`
	_, err = tx.Exec(ctx, updateHouseQuery, request.ID)
	if err != nil {
		return usecase.FlatResponse{}, err
	}

	err = repo.db.CommitTx(ctx, tx)
	if err != nil {
		return usecase.FlatResponse{}, err
	}

	return response, nil
}
