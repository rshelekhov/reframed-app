package handlers

import (
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/rshelekhov/remedi/internal/app"
	"github.com/rshelekhov/remedi/internal/lib/api/parser"
	resp "github.com/rshelekhov/remedi/internal/lib/api/response"
	"github.com/rshelekhov/remedi/internal/lib/logger/sl"
	"github.com/rshelekhov/remedi/internal/model"
	"github.com/rshelekhov/remedi/internal/storage"
	"log/slog"
	"net/http"
)

// NewHandler create a handler struct and register the routes
func newUserHandlers(r *chi.Mux, log *slog.Logger, app app.App, v *validator.Validate) {
	h := handler{
		logger:    log,
		app:       app,
		validator: v,
	}

	r.Post("/users", h.CreateUser())
	r.Get("/users/{id}", h.GetUser())
	r.Get("/users", h.GetUsers())
	r.Put("/users/{id}", h.UpdateUser())
	r.Delete("/users/{id}", h.DeleteUser())
	r.Get("/users/roles", h.GetUserRoles())
}

// CreateUser creates a new user
func (h *handler) CreateUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		ID     string `json:"id,omitempty"`
		RoleID int    `json:"role_id,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.CreateUser"

		log := sl.LogWithRequest(h.logger, op, r)

		user := &model.CreateUser{}

		// Decode the request body and validate the data
		err := DecodeAndValidate(w, r, log, user, h.validator)
		if err != nil {
			return
		}

		// Create the user
		id, err := h.app.CreateUser(user)
		// TODO: refactor to use a switch statement
		if err != nil {
			if errors.Is(err, storage.ErrUserAlreadyExists) {
				log.Error("user already exists", slog.String("email", user.Email))

				render.Status(r, http.StatusConflict)
				render.JSON(w, r, resp.Error("user already exists"))

				return
			}
			if errors.Is(err, storage.ErrRoleNotFound) {
				log.Error("role not found", slog.Int("role", user.RoleID))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, Response{
					Response: resp.Error("role not found"),
					RoleID:   user.RoleID,
				})

				return
			}
			log.Error("failed to create user", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to create user"))

			return
		}

		log.Info("User created", slog.Any("user_id", id))

		// Return the user id
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Response: resp.Success("User created"),
			ID:       id,
		})
	}
}

// GetUser get a user by ID
func (h *handler) GetUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		User model.GetUser `json:"user"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.GetUser"

		log := sl.LogWithRequest(h.logger, op, r)

		id, err := GetID(w, r, log)
		if err != nil {
			return
		}

		user, err := h.app.GetUser(id)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Error("user not found", slog.String("user_id", id))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("user not found"))

				return
			}
			log.Error("failed to get user", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get user"))

			return
		}

		log.Info("User received", slog.Any("user", user))

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response: resp.Success("User received"),
			User:     user,
		})
	}
}

// GetUsers get a list of users
func (h *handler) GetUsers() http.HandlerFunc {
	type Response struct {
		resp.Response
		Users []model.GetUser `json:"users"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.GetUsers"

		log := sl.LogWithRequest(h.logger, op, r)

		pagination, err := parser.ParseLimitAndOffset(r)
		if err != nil {
			log.Error("failed to parse limit and offset", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to parse limit and offset"))

			return
		}

		users, err := h.app.GetUsers(pagination)
		if err != nil {
			if errors.Is(err, storage.ErrNoUsersFound) {
				log.Error("no users found")

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("no users found"))

				return
			}
			log.Error("failed to get users", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get users"))

			return
		}

		log.Info(
			"users found",
			slog.Int("count", len(users)),
			slog.Int("limit", pagination.Limit),
			slog.Int("offset", pagination.Offset),
		)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response: resp.Success("users found"),
			Users:    users,
		})
	}
}

// UpdateUser updates a user by ID
func (h *handler) UpdateUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		ID    string `json:"id,omitempty"`
		Email string `json:"email,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.UpdateUser"

		log := sl.LogWithRequest(h.logger, op, r)

		user := &model.UpdateUser{}

		id, err := GetID(w, r, log)
		if err != nil {
			return
		}

		// Decode the request body and validate the data
		err = DecodeAndValidate(w, r, log, user, h.validator)
		if err != nil {
			return
		}

		err = h.app.UpdateUser(id, user)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Error("user not found", slog.String("user_id", id))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, Response{
					Response: resp.Error("user not found"),
					ID:       id,
				})

				return
			}
			if errors.Is(err, storage.ErrUserAlreadyExists) {
				log.Error("this email already taken", slog.String("email", user.Email))

				render.Status(r, http.StatusConflict)
				render.JSON(w, r, Response{
					Response: resp.Error("this email already taken"),
					Email:    user.Email,
				})

				return
			}
			log.Error("failed to update user", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to update user"))

			return
		}

		log.Info("User updated", slog.String("user_id", id))

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response: resp.Success("user updated"),
			ID:       id,
		})

	}
}

// DeleteUser deletes a user by ID
func (h *handler) DeleteUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		ID string `json:"id,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.DeleteUser"

		log := sl.LogWithRequest(h.logger, op, r)

		id, err := GetID(w, r, log)
		if err != nil {
			return
		}

		err = h.app.DeleteUser(id)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Error("user not found", slog.String("user_id", id))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, Response{
					Response: resp.Error("user not found"),
					ID:       id,
				})

				return
			}
			log.Error("failed to delete user", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to delete user"))

			return
		}

		log.Info("user deleted", slog.String("user_id", id))

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response: resp.Success("user deleted"),
			ID:       id,
		})
	}
}

// GetUserRoles get a list of roles
func (h *handler) GetUserRoles() http.HandlerFunc {
	type Response struct {
		resp.Response
		Roles []model.GetRole `json:"roles"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.GetUserRoles"

		log := sl.LogWithRequest(h.logger, op, r)

		roles, err := h.app.GetUserRoles()
		if err != nil {
			if errors.Is(err, storage.ErrNoRolesFound) {
				log.Error("no roles found")

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("no roles found"))

				return
			}
			log.Error("failed to get roles", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get roles"))

			return
		}

		log.Info("roles found", slog.Int("count", len(roles)))

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response: resp.Success("roles found"),
			Roles:    roles,
		})
	}
}
