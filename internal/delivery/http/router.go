package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Routes
	r.Get("/health", h.Health)

	r.Route("/events", func(r chi.Router) {
		r.Post("/", h.CreateEvent)
		r.Get("/", h.ListEvents)
		r.Get("/{id}", h.GetEvent)
	})

	return r
}
