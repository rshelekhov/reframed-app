package controller

import (
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"github.com/rshelekhov/reframed/internal/api/controller/parser"
	resp "github.com/rshelekhov/reframed/internal/api/controller/response"
	"github.com/rshelekhov/reframed/internal/logger"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/storage"
	"github.com/rshelekhov/reframed/internal/usecase"
	"log/slog"
	"net/http"
)

type UserController struct {
	Usecase usecase.User
	Logger  logger.Interface
}

// CreateUser creates a new user
func (c *UserController) CreateUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		ID     string `json:"id,omitempty"`
		RoleID int    `json:"role_id,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.CreateUser"

		log := logger.LogWithRequest(c.Logger, op, r)

		user := &model.User{}

		// Decode the request body
		err := DecodeJSON(w, r, log, user)
		if err != nil {
			return
		}

		fmt.Println(user)

		// Validate the request
		err = ValidateData(w, r, log, user)
		if err != nil {
			return
		}

		fmt.Println(user)

		// Create the user
		id, err := c.Usecase.CreateUser(r.Context(), user)
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Error("user already exists", slog.String("email", user.Email))

			render.Status(r, http.StatusConflict)
			render.JSON(w, r, resp.Error("user with this email already exists"))

			return
		}
		if err != nil {
			log.Error("failed to create user", logger.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to create user"))

			return
		}

		log.Info("User created", slog.Any("user_id", id))

		// Return the user id
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Response: resp.Success("user created"),
			ID:       id,
		})
	}
}

// GetUser get a user by ID
func (c *UserController) GetUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		User model.GetUser `json:"user"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.GetUser"

		log := logger.LogWithRequest(c.Logger, op, r)

		id, err := GetID(w, r, log)
		if err != nil {
			return
		}

		user, err := c.Usecase.GetUser(r.Context(), id)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", slog.String("user_id", id))

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("user not found"))

			return
		}
		if err != nil {
			log.Error("failed to get user", logger.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get user"))

			return
		}

		log.Info("User received", slog.Any("user", user))

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Response: resp.Success("user received"),
			User:     user,
		})
	}
}

// GetUsers get a list of users
func (c *UserController) GetUsers() http.HandlerFunc {
	type Response struct {
		resp.Response
		Users []*model.GetUser `json:"users"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.GetUsers"

		log := logger.LogWithRequest(c.Logger, op, r)

		pagination, err := parser.ParseLimitAndOffset(r)
		if err != nil {
			log.Error("failed to parse limit and offset", logger.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to parse limit and offset"))

			return
		}

		users, err := c.Usecase.GetUsers(r.Context(), pagination)
		if errors.Is(err, storage.ErrNoUsersFound) {
			log.Error("no users found")

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("no users found"))

			return
		}
		if err != nil {
			log.Error("failed to get users", logger.Err(err))

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
func (c *UserController) UpdateUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		ID    string `json:"id,omitempty"`
		Email string `json:"email,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.UpdateUser"

		log := logger.LogWithRequest(c.Logger, op, r)

		user := &model.UpdateUser{}

		id, err := GetID(w, r, log)
		if err != nil {
			return
		}

		// Decode the request body
		err = DecodeJSON(w, r, log, user)
		if err != nil {
			return
		}

		// Validate the request
		err = ValidateData(w, r, log, user)
		if err != nil {
			return
		}

		err = c.Usecase.UpdateUser(r.Context(), id, user)
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
		if err != nil {
			log.Error("failed to update user", logger.Err(err))

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
func (c *UserController) DeleteUser() http.HandlerFunc {
	type Response struct {
		resp.Response
		ID string `json:"id,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.DeleteUser"

		log := logger.LogWithRequest(c.Logger, op, r)

		id, err := GetID(w, r, log)
		if err != nil {
			return
		}

		err = c.Usecase.DeleteUser(r.Context(), id)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", slog.String("user_id", id))

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, Response{
				Response: resp.Error("user not found"),
				ID:       id,
			})

			return
		}
		if err != nil {
			log.Error("failed to delete user", logger.Err(err))

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
