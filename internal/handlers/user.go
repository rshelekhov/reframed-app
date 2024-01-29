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
	"strings"
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

// CreateUser creates a new user
func (h *UserHandler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.CreateUser"

		log := logger.LogWithRequest(h.Logger, op, r)

		user := &models.User{}

		// TODO: move DecodeJSON and ValidateData to another function
		// Decode the request body
		if err := DecodeJSON(w, r, log, user); err != nil {
			return
		}

		// Validate the request
		if err := ValidateData(w, r, log, user); err != nil {
			return
		}

		userID := ksuid.New().String()
		now := time.Now().UTC()

		newUser := models.User{
			ID:        userID,
			Email:     user.Email,
			Password:  user.Password,
			UpdatedAt: &now,
		}

		// Create the user
		err := h.UserStorage.CreateUser(r.Context(), newUser)
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

		// Register user device
		device, err := h.registerDevice(r, newUser.ID)
		if err != nil {
			log.Error(fmt.Sprintf("%v", ErrFailedToRegisterDevice), logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToRegisterDevice))
			return
		}

		// Create session
		tokens, expiresAt, err := h.createSession(r.Context(), userID, device.ID, h.TokenAuth.RefreshTokenTTL)
		if err != nil {
			log.Error(fmt.Sprintf("%v", ErrFailedToCreateSession), logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToCreateSession))
			return
		}

		additionalFields := map[string]string{"userID": userID}
		tokenData := auth.TokenData{
			AccessToken:      tokens.AccessToken,
			RefreshToken:     tokens.RefreshToken,
			Path:             r.URL.Path,
			ExpiresAt:        expiresAt,
			HttpOnly:         true,
			AdditionalFields: additionalFields,
		}

		log.Info(
			"user and tokens created",
			slog.String("user_id", userID),
			slog.Any("tokens", tokens),
		)
		auth.SendTokensToWeb(w, tokenData)
	}
}

func (h *UserHandler) LoginWithPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.LoginWithPassword"

		log := logger.LogWithRequest(h.Logger, op, r)

		userInput := &models.User{}

		// Decode the request body
		if err := DecodeJSON(w, r, log, userInput); err != nil {
			return
		}

		// Validate the request
		if err := ValidateData(w, r, log, userInput); err != nil {
			return
		}

		userDB, err := h.UserStorage.GetUserCredentials(r.Context(), userInput)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error(fmt.Sprintf("%v", ErrInvalidCredentials), slog.String("email", userInput.Email))
			responseError(w, r, http.StatusUnauthorized, fmt.Sprintf("%v", ErrInvalidCredentials))
			return
		}
		if err != nil {
			log.Error(fmt.Sprintf("%v", ErrFailedToGetData), logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToGetData))
			return
		}

		// TODO: add validate the password using bcrypt

		// Check if device exists (if not, register it)
		device, err := h.checkDevice(r, userDB.ID)
		if errors.Is(err, storage.ErrUserDeviceNotFound) {
			device, err = h.registerDevice(r, userDB.ID)
			if err != nil {
				log.Error(fmt.Sprintf("%v", ErrFailedToRegisterDevice), logger.Err(err))
				responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToRegisterDevice))
				return
			}
		}
		if err != nil {
			log.Error(fmt.Sprintf("%v", ErrFailedToCheckDevice), logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToCheckDevice))
			return
		}

		// Create session
		tokens, expiresAt, err := h.createSession(r.Context(), userDB.ID, device.ID, h.TokenAuth.RefreshTokenTTL)
		if err != nil {
			log.Error(fmt.Sprintf("%v", ErrFailedToCreateSession), logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToCreateSession))
			return
		}

		additionalFields := map[string]string{"userID": userDB.ID}
		tokenData := auth.TokenData{
			AccessToken:      tokens.AccessToken,
			RefreshToken:     tokens.RefreshToken,
			Path:             r.URL.Path,
			ExpiresAt:        expiresAt,
			HttpOnly:         true,
			AdditionalFields: additionalFields,
		}

		log.Info(
			"user logged in, tokens created",
			slog.String("user_id", userDB.ID),
			slog.Any("tokens", tokens),
		)
		auth.SendTokensToWeb(w, tokenData)
	}
}

// TODO: refactor it when we move auth to grpc (use as a reference Aooth)
func (h *UserHandler) RefreshJWTTokens() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.RefreshJWTTokens"

		log := logger.LogWithRequest(h.Logger, op, r)

		refreshToken, err := auth.FindRefreshToken(r)
		if err != nil {
			log.Error(fmt.Sprintf("%s: %v", op, ErrFailedToGetRefreshToken), logger.Err(err))
			responseError(w, r, http.StatusUnauthorized, fmt.Sprintf("%v: ", ErrFailedToGetRefreshToken))
			return
		}

		// Get session by refresh token
		session, err := h.UserStorage.GetSessionByRefreshToken(r.Context(), refreshToken)
		if errors.Is(err, storage.ErrSessionNotFound) {
			log.Error(
				fmt.Sprintf("%v", storage.ErrSessionNotFound),
				slog.String("refresh_token", refreshToken))
			responseError(w, r, http.StatusUnauthorized, fmt.Sprintf("%v", storage.ErrSessionNotFound))
			return
		}
		if errors.Is(err, storage.ErrRefreshTokenExpired) {
			log.Error(
				fmt.Sprintf("%v", storage.ErrRefreshTokenExpired),
				slog.String("user_id", session.UserID),
				slog.String("refresh_token", refreshToken),
			)
			responseError(w, r, http.StatusUnauthorized, fmt.Sprintf("%v", storage.ErrRefreshTokenExpired))
			return
		}
		if err != nil {
			log.Error(fmt.Sprintf("%v", ErrFailedToGetData), logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToGetData))
			return
		}

		// Check if refresh token is expired
		if session.ExpiresAt.Before(time.Now()) {
			log.Error(
				fmt.Sprintf("%v", storage.ErrRefreshTokenExpired),
				slog.String("user_id", session.UserID))
			responseError(w, r, http.StatusUnauthorized, fmt.Sprintf("%v", storage.ErrRefreshTokenExpired))
			return
		}

		// Check if device exists (if not, register it)
		device, err := h.checkDevice(r, session.UserID)
		if err != nil {
			log.Error(fmt.Sprintf("%v", ErrDeviceNotFound), logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrDeviceNotFound))
			return
		}

		// Create new tokens
		tokens, expiresAt, err := h.createSession(r.Context(), session.UserID, device.ID, h.TokenAuth.RefreshTokenTTL)
		if err != nil {
			log.Error(fmt.Sprintf("%v", ErrFailedToCreateSession), logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, fmt.Sprintf("%v", ErrFailedToCreateSession))
			return
		}

		tokenData := auth.TokenData{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			Path:         h.TokenAuth.RefreshTokenCookiePath,
			ExpiresAt:    expiresAt,
			HttpOnly:     true,
		}

		log.Info("tokens created", slog.Any("tokens", tokens))
		auth.SendTokensToWeb(w, tokenData)
	}
}

func (h *UserHandler) checkDevice(r *http.Request, userID string) (models.UserDevice, error) {

	device, err := h.UserStorage.GetUserDevice(r.Context(), userID, r.UserAgent())
	if errors.Is(err, storage.ErrUserDeviceNotFound) {
		return models.UserDevice{}, storage.ErrUserDeviceNotFound
	}
	if err != nil {
		return models.UserDevice{}, err
	}

	return device, nil
}

func (h *UserHandler) registerDevice(r *http.Request, userID string) (models.UserDevice, error) {
	ip := r.RemoteAddr
	ip = strings.Split(ip, ":")[0]

	latestLoginAt := time.Now()

	device := models.UserDevice{
		ID:            ksuid.New().String(),
		UserID:        userID,
		UserAgent:     r.UserAgent(),
		IP:            ip,
		Detached:      false,
		LatestLoginAt: &latestLoginAt,
		DetachedAt:    nil,
	}

	err := h.UserStorage.AddDevice(r.Context(), device)
	if err != nil {
		return models.UserDevice{}, err
	}

	return device, nil
}

// TODO: Move sessions from Postgres to Redis
func (h *UserHandler) createSession(
	ctx context.Context,
	userID, deviceID string,
	refreshTokenTTL time.Duration,
) (models.TokenResponse, time.Time, error) {
	var (
		resp models.TokenResponse
		err  error
	)

	additionalClaims := map[string]interface{}{
		"user_id": userID,
	}

	// TODO: move out from this function to LoginWithPassword function
	resp.AccessToken, err = h.TokenAuth.NewAccessToken(additionalClaims)
	if err != nil {
		return resp, time.Time{}, err
	}

	resp.RefreshToken, err = h.TokenAuth.NewRefreshToken()
	if err != nil {
		return resp, time.Time{}, err
	}

	expiresAt := time.Now().Add(refreshTokenTTL)

	session := models.Session{
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    &expiresAt,
	}

	err = h.UserStorage.SaveSession(ctx, userID, deviceID, session)
	if err != nil {
		return resp, time.Time{}, err
	}

	return resp, expiresAt, nil
}

// Logout
// TODO: add logout
func (h *UserHandler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.Logout"
	}
}

// GetUserByID get a user by ID
func (h *UserHandler) GetUserByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			op  = "user.handlers.GetUserByID"
			key = "userID"
		)

		log := logger.LogWithRequest(h.Logger, op, r)

		// TODO: get id from access token
		id, statusCode, err := GetID(r, log, key)
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
		const (
			op  = "user.handlers.UpdateUser"
			key = "userID"
		)

		log := logger.LogWithRequest(h.Logger, op, r)

		user := &models.UpdateUser{}

		// TODO: get id from access token
		id, statusCode, err := GetID(r, log, key)
		if err != nil {
			responseError(w, r, statusCode, err.Error())
			return
		}

		// Decode the request body
		if err = DecodeJSON(w, r, log, user); err != nil {
			return
		}

		// Validate the request
		if err = ValidateData(w, r, log, user); err != nil {
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
		const (
			op  = "user.handlers.DeleteUser"
			key = "userID"
		)

		log := logger.LogWithRequest(h.Logger, op, r)

		// TODO: get id from access token
		id, statusCode, err := GetID(r, log, key)
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
