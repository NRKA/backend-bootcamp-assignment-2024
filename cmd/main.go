package main

import (
	"github.com/NRKA/backend-bootcamp-assignment-2024/configs"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/app"
	"log"
)

func main() {
	config, err := configs.FromEnv()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	app.Run(config)
}
