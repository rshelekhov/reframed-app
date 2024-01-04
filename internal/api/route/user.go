package route

import (
	"github.com/go-chi/chi"
	"github.com/rshelekhov/reframed/internal/api/controller"
	"github.com/rshelekhov/reframed/internal/usecase"
	"github.com/rshelekhov/reframed/internal/usecase/storage"
	"log/slog"
)

// NewUserRouter create a handler struct and register the routes
func NewUserRouter(r *chi.Mux, log *slog.Logger, us *storage.UserStorage) {
	c := &controller.UserController{
		Usecase: usecase.NewUserUsecase(us),
		Logger:  log,
	}

	r.Route("/users", func(r chi.Router) {
		r.Post("/", c.CreateUser())
		r.Get("/", c.GetUsers())
		r.Get("/roles", c.GetUserRoles())

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", c.GetUser())
			r.Put("/", c.UpdateUser())
			r.Delete("/", c.DeleteUser())
		})
	})
}
