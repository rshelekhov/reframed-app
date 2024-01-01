package handlers

import (
	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
	"github.com/jmoiron/sqlx"
	"github.com/rshelekhov/remedi/internal/app"
	"github.com/rshelekhov/remedi/internal/storage/postgres"
	"log/slog"
)

type handler struct {
	logger    *slog.Logger
	app       app.App
	validator *validator.Validate
}

// Activate activates the user resource
func Activate(r *chi.Mux, log *slog.Logger, db *sqlx.DB, v *validator.Validate) {
	srv := app.New(postgres.GetStorage(db))
	newUserHandlers(r, log, srv, v)
}
