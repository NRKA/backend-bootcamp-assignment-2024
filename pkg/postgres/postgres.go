package postgres

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type Database struct {
	cluster *pgxpool.Pool
}

func NewDB(ctx context.Context, dbConfig DatabaseConfig) (*Database, error) {
	pool, err := pgxpool.New(ctx, GenerateDsn(dbConfig))
	if err != nil {
		return nil, err
	}

	return &Database{cluster: pool}, nil
}

func GenerateDsn(dbConfig DatabaseConfig) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Name)
}

func (db Database) GetPool() *pgxpool.Pool {
	return db.cluster
}

func (db Database) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.cluster, dest, query, args...)
}

func (db Database) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, db.cluster, dest, query, args...)
}

func (db Database) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.cluster.Exec(ctx, query, args...)
}

func (db Database) ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.cluster.QueryRow(ctx, query, args...)
}

func (db Database) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return db.cluster.Begin(ctx)
}

func (db Database) CommitTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

func (db Database) RollbackTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}
