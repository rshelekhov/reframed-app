package route

import (
	"github.com/go-chi/chi"
	"github.com/rshelekhov/reframed/internal/api/controller"
	"github.com/rshelekhov/reframed/internal/usecase"
	"github.com/rshelekhov/reframed/pkg/logger"
)

// NewUserRouter create a handler struct and register the routes
func NewUserRouter(r *chi.Mux, log logger.Interface, u usecase.User) {
	c := &controller.UserController{
		Usecase: u,
		Logger:  log,
	}

	r.Route("/users", func(r chi.Router) {
		r.Post("/", c.CreateUser())
		r.Get("/", c.GetUsers())

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", c.GetUser())
			r.Put("/", c.UpdateUser())
			r.Delete("/", c.DeleteUser())
		})
	})
}
