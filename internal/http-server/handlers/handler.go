package handlers

import (
	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"github.com/jmoiron/sqlx"
	"github.com/rshelekhov/reframed/internal/service"
	"github.com/rshelekhov/reframed/internal/storage/postgres"
	"log/slog"
)

type handler struct {
	logger    *slog.Logger
	srv       service.Service
	validator *validator.Validate
}

// Activate activates the user resource
func Activate(r *chi.Mux, log *slog.Logger, db *sqlx.DB, v *validator.Validate) {
	srv := service.New(postgres.GetStorage(db), v)
	newUserHandlers(r, log, srv, v)
}
