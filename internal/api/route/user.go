package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/rshelekhov/reframed/internal/api/handlers"
	"github.com/rshelekhov/reframed/internal/logger"
	"github.com/rshelekhov/reframed/internal/storage"
)

// NewUserRouter create a handlers struct and register the routes
func NewUserRouter(r *chi.Mux, log logger.Interface, s storage.UserStorage) {
	h := &handlers.UserHandler{
		Storage: s,
		Logger:  log,
	}

	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.CreateUser())
		r.Get("/", h.GetUsers())

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetUserByID())
			r.Put("/", h.UpdateUser())
			r.Delete("/", h.DeleteUser())
		})
	})
}
