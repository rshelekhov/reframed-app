package user

import (
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/jmoiron/sqlx"
	"github.com/rshelekhov/remedi/internal/lib/api/parser"
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
func Activate(r *chi.Mux, log *slog.Logger, db *sqlx.DB, validate *validator.Validate) {
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

	r.Post("/users", h.CreateUser())
	r.Get("/users/{id}", h.GetUser())
	r.Get("/users", h.GetUsers())
	r.Put("/users/{id}", h.UpdateUser())
	r.Delete("/users/{id}", h.DeleteUser())
	// TODO: add get user status
}

// CreateUser creates a new user
func (h *handler) CreateUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		ID     string `json:"id,omitempty"`
		RoleID int    `json:"role_id,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handler.CreateUser"

		log := sl.LogWithRequest(h.logger, op, r)

		var user CreateUser

		// Decode the request body
		err := render.DecodeJSON(r.Body, &user)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error(http.StatusNotFound, "request body is empty"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error(http.StatusBadRequest, "failed to decode request body"))

			return
		}

		log.Info("request body decoded", slog.Any("user", user))

		// Validate the user
		err = h.validator.Struct(user)
		if err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("failed to validate user", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		// Create the user
		id, err := h.service.CreateUser(user)
		if err != nil {
			if errors.Is(err, storage.ErrUserAlreadyExists) {
				log.Error("user already exists", slog.String("email", user.Email))

				render.JSON(w, r, resp.Error(http.StatusConflict, "user already exists"))

				return
			} else if errors.Is(err, storage.ErrRoleNotFound) {
				log.Error("role not found", slog.Int("role", user.RoleID))

				render.JSON(w, r, Response{
					Response: resp.Error(http.StatusNotFound, "role not found"),
					RoleID:   user.RoleID,
				})
			}
			log.Error("failed to create user", sl.Err(err))

			render.JSON(w, r, resp.Error(http.StatusInternalServerError, "failed to create user"))

			return
		}

		log.Info("User created", slog.Any("user_id", id))

		// Return the user id
		render.JSON(w, r, Response{
			Response: resp.Success(http.StatusCreated, "User created"),
			ID:       id,
		})
	}
}

// GetUser get a user by ID
func (h *handler) GetUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		User GetUser `json:"user"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handler.GetUser"

		log := sl.LogWithRequest(h.logger, op, r)

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Error("user id is empty")

			render.JSON(w, r, resp.Error(http.StatusBadRequest, "user id is empty"))

			return
		}

		user, err := h.service.GetUser(id)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Error("user not found", slog.String("user_id", id))

				render.JSON(w, r, resp.Error(http.StatusNotFound, "user not found"))

				return
			}
			log.Error("failed to get user", sl.Err(err))

			render.JSON(w, r, resp.Error(http.StatusInternalServerError, "failed to get user"))

			return
		}

		log.Info("User received", slog.Any("user", user))

		render.JSON(w, r, Response{
			Response: resp.Success(http.StatusOK, "User received"),
			User:     user,
		})
	}
}

// GetUsers get a list of users
func (h *handler) GetUsers() http.HandlerFunc {
	type Response struct {
		resp.Response
		Users []GetUser `json:"users"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handler.GetUsers"

		log := sl.LogWithRequest(h.logger, op, r)

		pagination, err := parser.ParseLimitAndOffset(r)
		if err != nil {
			log.Error("failed to parse limit and offset", sl.Err(err))

			render.JSON(w, r, resp.Error(http.StatusBadRequest, "failed to parse limit and offset"))

			return
		}

		users, err := h.service.GetUsers(pagination)
		if err != nil {
			if errors.Is(err, storage.ErrNoUsersFound) {
				log.Error("no users found")

				render.JSON(w, r, resp.Error(http.StatusNotFound, "no users found"))

				return
			}
			log.Error("failed to get users", sl.Err(err))

			render.JSON(w, r, resp.Error(http.StatusInternalServerError, "failed to get users"))
		}

		log.Info(
			"users found",
			slog.Int("count", len(users)),
			slog.Int("limit", pagination.Limit),
			slog.Int("offset", pagination.Offset),
		)

		render.JSON(w, r, Response{
			Response: resp.Success(http.StatusOK, "users found"),
			Users:    users,
		})
	}
}

// UpdateUser updates a user by ID
func (h *handler) UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handler.UpdateUser"
	}
}

// DeleteUser deletes a user by ID
func (h *handler) DeleteUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		ID string `json:"id,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handler.DeleteUser"

		log := sl.LogWithRequest(h.logger, op, r)

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Error("user id is empty")

			render.JSON(w, r, resp.Error(http.StatusBadRequest, "user id is empty"))

			return
		}

		err := h.service.DeleteUser(id)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Error("user not found", slog.String("user_id", id))

				render.JSON(w, r, Response{
					Response: resp.Error(http.StatusNotFound, "user not found"),
					ID:       id,
				})

				return
			}
			log.Error("failed to delete user", sl.Err(err))

			render.JSON(w, r, resp.Error(http.StatusInternalServerError, "failed to delete user"))

			return
		}

		log.Info("user deleted", slog.String("user_id", id))

		render.JSON(w, r, Response{
			Response: resp.Success(http.StatusOK, "user deleted"),
			ID:       id,
		})
	}
}
