package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/rshelekhov/reframed/internal/http-server/middleware/auth"
	"github.com/rshelekhov/reframed/internal/logger"
	"github.com/rshelekhov/reframed/internal/models"
	"github.com/rshelekhov/reframed/internal/storage"
	"github.com/segmentio/ksuid"
	"log/slog"
	"net/http"
	"time"
)

type UserHandler struct {
	Logger      logger.Interface
	TokenAuth   *auth.JWTAuth
	UserStorage storage.UserStorage
	ListStorage storage.ListStorage
}

func NewUserHandler(
	log logger.Interface,
	tokenAuth *auth.JWTAuth,
	userStorage storage.UserStorage,
	listStorage storage.ListStorage,
) *UserHandler {
	return &UserHandler{
		Logger:      log,
		TokenAuth:   tokenAuth,
		UserStorage: userStorage,
		ListStorage: listStorage,
	}
}

func (h *UserHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.Login"

		log := logger.LogWithRequest(h.Logger, op, r)

		userInput := &models.User{}

		// Decode the request body
		err := DecodeJSON(w, r, log, userInput)
		if err != nil {
			return
		}

		// Validate the request
		err = ValidateData(w, r, log, userInput)
		if err != nil {
			return
		}

		userDB, err := h.UserStorage.GetUserCredentials(r.Context(), userInput)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error(fmt.Sprintf("%v", storage.ErrUserNotFound), slog.String("email", userInput.Email))
			responseError(w, r, http.StatusNotFound, fmt.Sprintf("%v", storage.ErrUserNotFound))
			return
		}
		if err != nil {
			log.Error(fmt.Sprintf("%v", ErrFailedToGetData), logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToGetData))
			return
		}

		// TODO: add validate the password using bcrypt

		// Validate token
		/*if userInput.Email == userDB.Email && userInput.Password == userDB.Password {
			tokenString, err := auth.CreateToken(userDB.ID)
			if err != nil {
				log.Error(fmt.Sprintf("%v", ErrFailedToCreateToken), logger.Err(err))
				responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToCreateToken))
				return
			}

			log.Info("token created", slog.String("token", tokenString))
			responseSuccess(w, r, http.StatusOK, "token created", models.TokenResponse{AccessToken: tokenString})
			return
		} else {
			log.Error(fmt.Sprintf("%v", ErrInvalidCredentials), slog.String("email", userInput.Email))
			responseError(w, r, http.StatusUnauthorized, fmt.Sprintf("%v", ErrInvalidCredentials))
		}*/
	}
}

func (h *UserHandler) RefreshTokens() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.RefreshTokens"

		h.createSession(r.Context(), user.ID, h.TokenAuth.AccessTokenTTL, h.TokenAuth.RefreshTokenTTL)
	}
}

// TODO: add logout
func (h *UserHandler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.Logout"
	}
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
		err = h.UserStorage.CreateUser(r.Context(), newUser)
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Error(fmt.Sprintf("%v", storage.ErrUserAlreadyExists), slog.String("email", user.Email))
			responseError(w, r, http.StatusBadRequest, fmt.Sprintf("%v", storage.ErrUserAlreadyExists))
			return
		} else if err != nil {
			log.Error("failed to create user", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to create user")
			return
		}

		// Create "Inbox" list
		listID := ksuid.New().String()
		now = time.Now().UTC()

		newList := models.List{
			ID:        listID,
			Title:     "Inbox",
			UserID:    newUser.ID,
			UpdatedAt: &now,
		}

		err = h.ListStorage.CreateList(r.Context(), newList)
		if err != nil {
			log.Error("failed to create list", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to create list")
			// return
		}

		log.Info("user created", slog.Any("user_id", id))
		responseSuccess(w, r, http.StatusCreated, "user created", models.User{ID: id})
	}
}

func (h *UserHandler) createSession(
	ctx context.Context,
	userID string,
	accessTokenTTL,
	refreshTokenTTL time.Duration,
) (models.TokenResponse, error) {
	var (
		resp models.TokenResponse
		err  error
	)

	resp.AccessToken, err = h.TokenAuth.CreateToken(userID, accessTokenTTL)
	if err != nil {
		return resp, err
	}

	resp.RefreshToken, err = h.TokenAuth.NewRefreshToken()
	if err != nil {
		return resp, err
	}

	session := models.Session{
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(refreshTokenTTL), // TODO: move ttl to config and set 720h
	}

	err = h.UserStorage.SetSession(ctx, userID, session)
	if err != nil {
		return resp, err
	}

	return resp, nil
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

		user, err := h.UserStorage.GetUserByID(r.Context(), id)
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

		users, err := h.UserStorage.GetUsers(r.Context(), pagination)
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

		err = h.UserStorage.UpdateUser(r.Context(), updatedUser)
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

		err = h.UserStorage.DeleteUser(r.Context(), id)
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
