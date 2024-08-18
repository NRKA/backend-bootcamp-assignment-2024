package router

import (
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/app/handlers"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/auth"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/house"
	"github.com/go-chi/chi/v5"
)

func New(auth *auth.Repo, house *house.Repo) *chi.Mux {
	router := chi.NewRouter()

	//No auth routes
	router.Get("/dummyLogin", handlers.DummyLoginHandler)
	router.Post("/login", auth.Login)
	router.Post("/register", auth.Register)

	//Auth only routes
	router.Get("/house/{id}", handlers.HouseHandler)
	router.Post("/house/{id}/subscribe", handlers.SubscribeHouseHandler)
	router.Post("/flat/create", handlers.CreateFlatHandler)

	// moderation only
	router.Post("/house/create", house.Create)
	router.Post("/flat/update", handlers.UpdateFlatHandler)
	return router
}
