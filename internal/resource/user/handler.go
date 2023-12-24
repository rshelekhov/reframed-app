package handlers

import (
	"database/sql"
	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"github.com/rshelekhov/remedi/internal/resource/user/service"
	"github.com/rshelekhov/remedi/internal/resource/user/storage"
	"log/slog"
)

type handler struct {
	logger  *slog.Logger
	service service.Service
}

func Activate(r *chi.Mux, log *slog.Logger, db *sql.DB, validate *validator.Validate) {
	srv := service.NewService(validate, storage.NewStorage(db))
	newHandler(r, log, srv)
}

func newHandler(r *chi.Mux, log *slog.Logger, srv service.Service) {
	h := handler{
		logger:  log,
		service: srv,
	}

	r.Get("/users", h.ListUsers())
	r.Post("/users", h.CreateUser())
	r.Get("/users/{id}", h.ReadUser())
	r.Put("/users/{id}", h.UpdateUser())
	r.Delete("/users/{id}", h.DeleteUser())
}
