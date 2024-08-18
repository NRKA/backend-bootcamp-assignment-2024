package configs

import (
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"github.com/joho/godotenv"
	"os"
)

const (
	dbHost     = "DB_HOST"
	dbPort     = "DB_PORT"
	dbUser     = "DB_USER"
	dbPassword = "DB_PASSWORD"
	dbName     = "DB_NAME"
)

func FromEnv() postgres.DatabaseConfig {
	godotenv.Load()
	return postgres.DatabaseConfig{
		Host:     os.Getenv(dbHost),
		Port:     os.Getenv(dbPort),
		User:     os.Getenv(dbUser),
		Password: os.Getenv(dbPassword),
		Name:     os.Getenv(dbName),
	}
}
