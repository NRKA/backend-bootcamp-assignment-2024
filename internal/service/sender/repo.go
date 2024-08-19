package sender

import (
	"context"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/usecase"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
)

type Repo struct {
	db *postgres.Database
}

func NewRepo(db *postgres.Database) *Repo {
	return &Repo{
		db: db,
	}
}

func (repo *Repo) Subscribe(ctx context.Context, houseID int, subscriber usecase.Subscribe) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM house WHERE id = $1)`

	err := repo.db.Get(ctx, &exists, query, houseID)
	if err != nil {
		return err
	}

	if !exists {
		return ErrHouseNotFound
	}

	insertQuery := `
		INSERT INTO subscriber (email, house_id)
		VALUES ($1, $2)
		ON CONFLICT (house_id) DO NOTHING
	`

	_, err = repo.db.Exec(ctx, insertQuery, subscriber.Email, houseID)
	if err != nil {
		return err
	}

	return nil
}
