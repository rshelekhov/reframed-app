package handlers

import (
	"errors"
	"fmt"
	"github.com/rshelekhov/reframed/internal/logger"
	"github.com/rshelekhov/reframed/internal/models"
	"github.com/rshelekhov/reframed/internal/storage"
	"github.com/segmentio/ksuid"
	"log/slog"
	"net/http"
	"time"
)

type UserHandler struct {
	Storage storage.UserStorage
	Logger  logger.Interface
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.CreateUser"

		log := logger.LogWithRequest(h.Logger, op, r)

		user := &models.User{}

		// Decode the request body
		err := DecodeJSON(w, r, log, user)
		if err != nil {
			return
		}

		// Validate the request
		err = ValidateData(w, r, log, user)
		if err != nil {
			return
		}

		id := ksuid.New().String()
		now := time.Now().UTC()

		newUser := models.User{
			ID:        id,
			Email:     user.Email,
			Password:  user.Password,
			UpdatedAt: &now,
		}

		// Create the user
		err = h.Storage.CreateUser(r.Context(), newUser)
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Error(fmt.Sprintf("%v", storage.ErrUserAlreadyExists), slog.String("email", user.Email))
			responseError(w, r, http.StatusBadRequest, fmt.Sprintf("%v", storage.ErrUserAlreadyExists))
			return
		} else if err != nil {
			log.Error("failed to create user", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to create user")
			return
		} else {
			log.Info("user created", slog.Any("user_id", id))
			responseSuccess(w, r, http.StatusCreated, "user created", models.User{ID: id})
		}
	}
}

// GetUserByID get a user by ID
func (h *UserHandler) GetUserByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.GetUserByID"

		log := logger.LogWithRequest(h.Logger, op, r)

		id, statusCode, err := GetID(r, log)
		if err != nil {
			responseError(w, r, statusCode, err.Error())
			return
		}

		user, err := h.Storage.GetUserByID(r.Context(), id)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error(fmt.Sprintf("%v", storage.ErrUserNotFound), slog.String("user_id", id))
			responseError(w, r, http.StatusNotFound, fmt.Sprintf("%v", storage.ErrUserNotFound))
			return
		}
		if err != nil {
			log.Error(fmt.Sprintf("%v", ErrFailedToGetData), logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToGetData))
			return
		}

		log.Info("user received", slog.Any("user", user))
		responseSuccess(w, r, http.StatusOK, "user received", user)
	}
}

// GetUsers get a list of users
func (h *UserHandler) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.GetUsers"

		log := logger.LogWithRequest(h.Logger, op, r)

		pagination, err := ParseLimitAndOffset(r)
		if err != nil {
			log.Error(ErrFailedToParsePagination.Error(), logger.Err(err))
			responseError(w, r, http.StatusBadRequest, ErrFailedToParsePagination.Error())
			return
		}

		users, err := h.Storage.GetUsers(r.Context(), pagination)
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
func (h *UserHandler) UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.UpdateUser"

		log := logger.LogWithRequest(h.Logger, op, r)

		user := &models.UpdateUser{}

		id, statusCode, err := GetID(r, log)
		if err != nil {
			responseError(w, r, statusCode, err.Error())
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

		now := time.Now().UTC()

		updatedUser := models.User{
			ID:        id,
			Email:     user.Email,
			Password:  user.Password,
			UpdatedAt: &now,
		}

		err = h.Storage.UpdateUser(r.Context(), updatedUser)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error(fmt.Sprintf("%v", storage.ErrUserNotFound), slog.String("user_id", id))
			responseError(w, r, http.StatusNotFound, fmt.Sprintf("%v", storage.ErrUserNotFound))
			return
		}
		if errors.Is(err, storage.ErrEmailAlreadyTaken) {
			log.Error(fmt.Sprintf("%v", storage.ErrEmailAlreadyTaken), slog.String("email", user.Email))
			responseError(w, r, http.StatusBadRequest, fmt.Sprintf("%v", storage.ErrEmailAlreadyTaken))
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

		log.Info("user updated", slog.String("user_id", id))
		responseSuccess(w, r, http.StatusOK, "user updated", models.User{ID: id})
	}
}

// DeleteUser deletes a user by ID
func (h *UserHandler) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.DeleteUser"

		log := logger.LogWithRequest(h.Logger, op, r)

		id, statusCode, err := GetID(r, log)
		if err != nil {
			responseError(w, r, statusCode, err.Error())
			return
		}

		err = h.Storage.DeleteUser(r.Context(), id)
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

		responseSuccess(w, r, http.StatusOK, "user deleted", models.User{ID: id})
	}
}
