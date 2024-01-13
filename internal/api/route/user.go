package route

import (
	"github.com/go-chi/chi"
	"github.com/rshelekhov/reframed/internal/api/handler"
	"github.com/rshelekhov/reframed/internal/logger"
	"github.com/rshelekhov/reframed/internal/usecase"
)

// NewUserRouter create a handler struct and register the routes
func NewUserRouter(r *chi.Mux, log logger.Interface, u usecase.User) {
	c := &handler.UserHandler{
		Usecase: u,
		Logger:  log,
	}

	r.Route("/users", func(r chi.Router) {
		r.Post("/", c.CreateUser())
		r.Get("/", c.GetUsers())

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", c.GetUserByID())
			r.Put("/", c.UpdateUser())
			r.Delete("/", c.DeleteUser())
		})
	})
}
