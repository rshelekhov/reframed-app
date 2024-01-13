package controller

import (
	"errors"
	"fmt"
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
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.CreateUser"

		log := logger.LogWithRequest(c.Logger, op, r)

		// user := &model.CreateUser{}
		user := &model.User{}

		// Decode the request body
		err := decodeJSON(w, r, log, user)
		if err != nil {
			return
		}

		// Validate the request
		err = validateData(w, r, log, user)
		if err != nil {
			return
		}

		// Create the user
		id, err := c.Usecase.CreateUser(r.Context(), user)
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Error(fmt.Sprintf("%v", storage.ErrUserAlreadyExists), slog.String("email", *user.Email))
			responseError(w, r, http.StatusConflict, fmt.Sprintf("%v", storage.ErrUserAlreadyExists))
			return
		}
		if err != nil {
			log.Error("failed to create user", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to create user")
			return
		}

		log.Info("User created", slog.Any("user_id", id))
		responseSuccess(w, r, http.StatusCreated, "user created", model.User{ID: id})
	}
}

// GetUserByID get a user by ID
func (c *UserController) GetUserByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.GetUserByID"

		log := logger.LogWithRequest(c.Logger, op, r)

		id, err := GetID(w, r, log)
		if err != nil {
			return
		}

		user, err := c.Usecase.GetUserByID(r.Context(), id)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error(fmt.Sprintf("%v", storage.ErrUserNotFound), slog.String("user_id", id))
			responseError(w, r, http.StatusNotFound, fmt.Sprintf("%v", storage.ErrUserNotFound))
			return
		}
		if err != nil {
			log.Error("failed to get user", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to get user")
			return
		}

		log.Info("User received", slog.Any("user", user))
		responseSuccess(w, r, http.StatusOK, "user received", user)
	}
}

// GetUsers get a list of users
func (c *UserController) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.GetUsers"

		log := logger.LogWithRequest(c.Logger, op, r)

		pagination, err := parseLimitAndOffset(r)
		if err != nil {
			log.Error("failed to parse limit and offset", logger.Err(err))
			responseError(w, r, http.StatusBadRequest, "failed to parse limit and offset")
			return
		}

		users, err := c.Usecase.GetUsers(r.Context(), pagination)
		if errors.Is(err, storage.ErrNoUsersFound) {
			log.Error(fmt.Sprintf("%v", storage.ErrNoUsersFound))
			responseError(w, r, http.StatusNotFound, fmt.Sprintf("%v", storage.ErrNoUsersFound))
			return
		}
		if err != nil {
			log.Error("failed to get users", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to get users")
			return
		}

		log.Info(
			"users found",
			slog.Int("count", len(users)),
			slog.Int("limit", pagination.Limit),
			slog.Int("offset", pagination.Offset),
		)

		responseSuccess(w, r, http.StatusOK, "users found", users)
	}
}

// UpdateUser updates a user by ID
func (c *UserController) UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.UpdateUser"

		log := logger.LogWithRequest(c.Logger, op, r)

		user := &model.UpdateUser{}

		id, err := GetID(w, r, log)
		if err != nil {
			return
		}

		// Decode the request body
		err = decodeJSON(w, r, log, user)
		if err != nil {
			return
		}

		// Validate the request
		err = validateData(w, r, log, user)
		if err != nil {
			return
		}

		err = c.Usecase.UpdateUser(r.Context(), id, user)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error(fmt.Sprintf("%v", storage.ErrUserNotFound), slog.String("user_id", id))
			responseError(w, r, http.StatusNotFound, fmt.Sprintf("%v", storage.ErrUserNotFound))
			return
		}
		if errors.Is(err, storage.ErrEmailAlreadyTaken) {
			log.Error(fmt.Sprintf("%v", storage.ErrEmailAlreadyTaken), slog.String("email", user.Email))
			responseError(w, r, http.StatusConflict, fmt.Sprintf("%v", storage.ErrEmailAlreadyTaken))
			return
		}
		if errors.Is(err, storage.ErrNoChangesDetected) {
			log.Error(fmt.Sprintf("%v", storage.ErrNoChangesDetected), slog.String("user_id", id))
			responseError(w, r, http.StatusBadRequest, fmt.Sprintf("%v", storage.ErrNoChangesDetected))
			return
		}
		if errors.Is(err, storage.ErrNoPasswordChangesDetected) {
			log.Error(fmt.Sprintf("%v", storage.ErrNoPasswordChangesDetected), slog.String("user_id", id))
			responseError(w, r, http.StatusBadRequest, fmt.Sprintf("%v", storage.ErrNoPasswordChangesDetected))
			return
		}
		if err != nil {
			log.Error("failed to update user", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to update user")
			return
		}

		log.Info("User updated", slog.String("user_id", id))
		responseSuccess(w, r, http.StatusOK, "user updated", model.User{ID: id})
	}
}

// DeleteUser deletes a user by ID
func (c *UserController) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.DeleteUser"

		log := logger.LogWithRequest(c.Logger, op, r)

		id, err := GetID(w, r, log)
		if err != nil {
			return
		}

		err = c.Usecase.DeleteUser(r.Context(), id)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error(fmt.Sprintf("%v", storage.ErrUserNotFound), slog.String("user_id", id))
			responseError(w, r, http.StatusNotFound, fmt.Sprintf("%v", storage.ErrUserNotFound))
			return
		}
		if err != nil {
			log.Error("failed to delete user", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to delete user")
			return
		}

		log.Info("user deleted", slog.String("user_id", id))

		responseSuccess(w, r, http.StatusOK, "user deleted", model.User{ID: id})
	}
}
