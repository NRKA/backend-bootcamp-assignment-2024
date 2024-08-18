package main

import (
	"context"
	"github.com/NRKA/backend-bootcamp-assignment-2024/configs"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/app/router"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/auth"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/house"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"log"
	"net/http"
)

func main() {
	cfg := configs.FromEnv()

	db, err := postgres.NewDB(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}

	authRepo := auth.NewAuthRepo(db)
	houseRepo := house.NewHouseRepo(db)
	r := router.New(authRepo, houseRepo)
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
