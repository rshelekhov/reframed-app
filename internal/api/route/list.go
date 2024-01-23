package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/rshelekhov/reframed/internal/api/handlers"
	"github.com/rshelekhov/reframed/internal/logger"
	"github.com/rshelekhov/reframed/internal/storage"
)

func NewListRouter(r *chi.Mux, log logger.Interface, s storage.ListStorage) {
	h := &handlers.ListHandler{
		Storage: s,
		Logger:  log,
	}

	r.Route("/lists", func(r chi.Router) {
		r.Get("/", h.GetLists())
		r.Post("/", h.CreateList())

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetListByID())
			r.Put("/", h.UpdateList())
			r.Delete("/", h.DeleteList())
		})
	})
}
