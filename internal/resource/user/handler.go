package user

import (
	"database/sql"
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	resp "github.com/rshelekhov/remedi/internal/lib/api/response"
	"github.com/rshelekhov/remedi/internal/lib/logger/sl"
	"github.com/rshelekhov/remedi/internal/storage"
	"io"
	"log/slog"
	"net/http"
)

type handler struct {
	logger    *slog.Logger
	service   Service
	validator *validator.Validate
}

// Activate activates the user resource
func Activate(r *chi.Mux, log *slog.Logger, db *sql.DB, validate *validator.Validate) {
	srv := NewService(NewStorage(db))
	newHandler(r, log, srv, validate)
}

// NewHandler create a handler struct and register the routes
func newHandler(r *chi.Mux, log *slog.Logger, srv Service, validate *validator.Validate) {
	h := handler{
		logger:    log,
		service:   srv,
		validator: validate,
	}

	r.Get("/users", h.ListUsers())
	r.Post("/users", h.CreateUser())
	r.Get("/users/{id}", h.ReadUser())
	r.Put("/users/{id}", h.UpdateUser())
	r.Delete("/users/{id}", h.DeleteUser())
}

// ListUsers get a list users
func (h *handler) ListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.ListUsers"
	}
}

// CreateUser creates a new user
func (h *handler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.CreateUser"

		log := h.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var user CreateUser

		// Decode the request body
		err := render.DecodeJSON(r.Body, &user)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("request body is empty"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request body"))

			return
		}

		log.Info("request body decoded", slog.Any("user", user))

		// Validate the user
		err = h.validator.Struct(user)
		if err != nil {
			validateErr := err.(validator.ValidationErrors)

			render.JSON(w, r, resp.Error(validateErr.Error()))

			return
		}

		// Create the user
		id, err := h.service.CreateUser(user)
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Info("user already exists", slog.String("email", user.Email))

			render.JSON(w, r, resp.Error("user already exists"))

			return
		}
		if err != nil {
			log.Error("failed to create user", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to create user"))

			return
		}

		log.Info("User created", slog.Any("user_id", id))

		// Return the user id
		render.JSON(w, r, resp.Success("User created", id))
	}
}

// ReadUser get a user by id
func (h *handler) ReadUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.ReadUser"
	}
}

// UpdateUser updates a user by id
func (h *handler) UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.UpdateUser"
	}
}

// DeleteUser deletes a user by id
func (h *handler) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.DeleteUser"

		id := uuid.New()
		err := h.service.DeleteUser(id)
		if err != nil {
			return
		}
	}
}
