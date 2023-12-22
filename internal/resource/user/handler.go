package user

import (
	"database/sql"
	"log/slog"
	"net/http"
)

type API struct {
	logger  *slog.Logger
	storage *Storage
}

func New(log *slog.Logger, db *sql.DB) *API {
	return &API{
		logger:  log,
		storage: NewRepository(db),
	}
}

func (a *API) ListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {

	}
}

func (a *API) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// user := models.CreateUser{}
	}
}

func (a *API) ReadUser() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {

	}
}

func (a *API) UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {

	}
}

func (a *API) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {

	}
}
