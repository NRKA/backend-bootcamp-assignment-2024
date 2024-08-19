package app

import (
	"context"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/server/router"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/auth"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/flat"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/house"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/sender"
	"github.com/NRKA/backend-bootcamp-assignment-2024/pkg/postgres"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

const port = "PORT"

func Run(config postgres.DatabaseConfig) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	db, err := postgres.NewDB(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	defer db.GetPool().Close()

	auth := auth.NewHandler(db)
	house := house.NewHandler(db)
	flat := flat.NewHandler(db)
	s := sender.NewHandler(db)
	r := router.New(auth, house, flat, s)

	port := os.Getenv(port)
	var wg sync.WaitGroup
	wg.Add(2)

	sender := sender.New()
	go func() {
		defer wg.Done()

		if err := sender.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()

		if err = http.ListenAndServe(port, r); err != nil {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	wg.Wait()
}
