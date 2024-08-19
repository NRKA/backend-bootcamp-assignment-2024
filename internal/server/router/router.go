package router

import (
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/server/middleware"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/auth"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/flat"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/house"
	"github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/sender"
	"github.com/go-chi/chi/v5"
)

func New(auth *auth.Handler, house *house.Handler, flat *flat.Handler, sender *sender.Handler) *chi.Mux {
	router := chi.NewRouter()

	// No auth
	router.Get("/dummyLogin", auth.DummyLogin)
	router.Post("/login", auth.Login)
	router.Post("/register", auth.Register)

	// Auth only
	router.Group(func(r chi.Router) {
		r.Use(middleware.TokenAuthenticator, middleware.AuthOnly)
		r.Get("/house/{id}", house.Flats)
		r.Post("/house/{id}/subscribe", sender.Subscribe)
		r.Post("/flat/create", flat.Create)
	})

	// Moderation only
	router.Group(func(r chi.Router) {
		r.Use(middleware.TokenAuthenticator, middleware.ModerationOnly)
		r.Post("/house/create", house.Create)
		r.Post("/flat/update", flat.Update)
	})

	return router
}
