package handlers

import (
	"context"
	"errors"
	"github.com/rshelekhov/reframed/src/le"
	"github.com/rshelekhov/reframed/src/logger"
	"github.com/rshelekhov/reframed/src/models"
	"github.com/rshelekhov/reframed/src/server/middleware/jwtoken"
	"github.com/rshelekhov/reframed/src/storage"
	"github.com/segmentio/ksuid"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type UserHandler struct {
	Logger      logger.Interface
	TokenAuth   *jwtoken.JWTAuth
	UserStorage storage.UserStorage
	ListStorage storage.ListStorage
}

func NewUserHandler(
	log logger.Interface,
	tokenAuth *jwtoken.JWTAuth,
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

		/*
			// Decode the request body
			if err = DecodeJSON(w, r, log, user); err != nil {
				return
			}

			// Validate the request
			if err = ValidateData(w, r, log, user); err != nil {
				return
			}
		*/

		if err := DecodeAndValidateJSON(w, r, log, user); err != nil {
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
		if errors.Is(err, le.ErrUserAlreadyExists) {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrUserAlreadyExists, slog.String("email", user.Email))
			return
		} else if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToCreateUser, err)
			return
		}

		/*
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
				handleInternalServerError(w, r, log, le.ErrFailedToCreateList, err)
				return
			}
		*/

		// Register user device
		device, err := h.registerDevice(r, newUser.ID)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToRegisterDevice, err)
			return
		}

		// Create session
		tokens, expiresAt, err := h.createSession(r.Context(), userID, device.ID, h.TokenAuth.RefreshTokenTTL)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToCreateSession, err)
			return
		}

		additionalFields := map[string]string{"userID": userID}
		tokenData := jwtoken.TokenData{
			AccessToken:      tokens.AccessToken,
			RefreshToken:     tokens.RefreshToken,
			Domain:           h.TokenAuth.RefreshTokenCookieDomain,
			Path:             h.TokenAuth.RefreshTokenCookiePath,
			ExpiresAt:        expiresAt,
			HttpOnly:         true,
			AdditionalFields: additionalFields,
		}

		log.Info(
			"user and tokens created",
			slog.String("user_id", userID),
			slog.Any("tokens", tokens),
		)
		jwtoken.SendTokensToWeb(w, tokenData)
	}
}

func (h *UserHandler) LoginWithPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.LoginWithPassword"

		log := logger.LogWithRequest(h.Logger, op, r)

		userInput := &models.User{}

		/*
			// Decode the request body
			if err := DecodeJSON(w, r, log, userInput); err != nil {
				return
			}

			// Validate the request
			if err := ValidateData(w, r, log, userInput); err != nil {
				return
			}
		*/

		if err := DecodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		userDB, err := h.UserStorage.GetUserCredentials(r.Context(), userInput)
		if errors.Is(err, le.ErrUserNotFound) {
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrInvalidCredentials, slog.String("email", userInput.Email))
			return
		}
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		}

		// TODO: add validate the password using bcrypt

		// Check if device exists (if not, register it)
		device, err := h.checkDevice(r, userDB.ID)
		if errors.Is(err, le.ErrUserDeviceNotFound) {
			device, err = h.registerDevice(r, userDB.ID)
			if err != nil {
				handleInternalServerError(w, r, log, le.ErrFailedToRegisterDevice, err)
				return
			}
		}
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToCheckDevice, err)
			return
		}

		// Create session
		tokens, expiresAt, err := h.createSession(r.Context(), userDB.ID, device.ID, h.TokenAuth.RefreshTokenTTL)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToCreateSession, err)
			return
		}

		additionalFields := map[string]string{"userID": userDB.ID}
		tokenData := jwtoken.TokenData{
			AccessToken:      tokens.AccessToken,
			RefreshToken:     tokens.RefreshToken,
			Domain:           h.TokenAuth.RefreshTokenCookieDomain,
			Path:             h.TokenAuth.RefreshTokenCookiePath,
			ExpiresAt:        expiresAt,
			HttpOnly:         true,
			AdditionalFields: additionalFields,
		}

		log.Info(
			"user logged in, tokens created",
			slog.String("user_id", userDB.ID),
			slog.Any("tokens", tokens),
		)
		jwtoken.SendTokensToWeb(w, tokenData)
	}
}

// TODO: refactor it when we move jwtoken to grpc (use as a reference Aooth)
func (h *UserHandler) RefreshJWTTokens() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.RefreshJWTTokens"

		log := logger.LogWithRequest(h.Logger, op, r)

		refreshToken, err := jwtoken.FindRefreshToken(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrFailedToGetRefreshToken, err)
			return
		}

		// Get session by refresh token
		session, err := h.UserStorage.GetSessionByRefreshToken(r.Context(), refreshToken)
		if errors.Is(err, le.ErrSessionNotFound) {
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrSessionNotFound, slog.String("refresh_token", refreshToken))
			return
		}
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		}

		// Check if refresh token is expired
		if session.ExpiresAt.Before(time.Now()) {
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrRefreshTokenExpired, slog.String("user_id", session.UserID))
			return
		}

		// Check if device exists (if not, register it)
		device, err := h.checkDevice(r, session.UserID)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrUserDeviceNotFound, err)
			return
		}

		// Create new tokens
		tokens, expiresAt, err := h.createSession(r.Context(), session.UserID, device.ID, h.TokenAuth.RefreshTokenTTL)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToCreateSession, err)
			return
		}

		tokenData := jwtoken.TokenData{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			Domain:       h.TokenAuth.RefreshTokenCookieDomain,
			Path:         h.TokenAuth.RefreshTokenCookiePath,
			ExpiresAt:    expiresAt,
			HttpOnly:     true,
		}

		log.Info("tokens created", slog.Any("tokens", tokens))
		jwtoken.SendTokensToWeb(w, tokenData)
	}
}

func (h *UserHandler) checkDevice(r *http.Request, userID string) (models.UserDevice, error) {

	device, err := h.UserStorage.GetUserDevice(r.Context(), userID, r.UserAgent())
	if errors.Is(err, le.ErrUserDeviceNotFound) {
		return models.UserDevice{}, le.ErrUserDeviceNotFound
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
		contextUserID: userID,
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

// GetUser get a user by ID
func (h *UserHandler) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handlers.GetUser"

		log := logger.LogWithRequest(h.Logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetAccessToken, err)
			return
		}
		id := claims[contextUserID].(string)

		user, err := h.UserStorage.GetUser(r.Context(), id)
		if errors.Is(err, le.ErrUserNotFound) {
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound, slog.String("user_id", id))
			return
		}
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
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
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrFailedToParsePagination, err)
			return
		}

		users, err := h.UserStorage.GetUsers(r.Context(), pagination)
		if errors.Is(err, le.ErrNoUsersFound) {
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoUsersFound)
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

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetAccessToken, err)
			return
		}
		id := claims[contextUserID].(string)

		/*
			// Decode the request body
			if err = DecodeJSON(w, r, log, user); err != nil {
				return
			}

			// Validate the request
			if err = ValidateData(w, r, log, user); err != nil {
				return
			}
		*/

		if err = DecodeAndValidateJSON(w, r, log, user); err != nil {
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
		if errors.Is(err, le.ErrUserNotFound) {
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound, slog.String("user_id", id))
			return
		}
		if errors.Is(err, le.ErrEmailAlreadyTaken) {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmailAlreadyTaken, slog.String("email", user.Email))
			return
		}
		if errors.Is(err, le.ErrNoChangesDetected) {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrNoChangesDetected, slog.String("user_id", id))
			return
		}
		if errors.Is(err, le.ErrNoPasswordChangesDetected) {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrNoPasswordChangesDetected, slog.String("user_id", id))
			return
		}
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateUser, slog.String("user_id", id))
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

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetAccessToken, err)
			return
		}
		id := claims[contextUserID].(string)

		err = h.UserStorage.DeleteUser(r.Context(), id)
		if errors.Is(err, le.ErrUserNotFound) {
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound, slog.String("user_id", id))
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
