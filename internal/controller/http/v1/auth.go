package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type authController struct {
	logger  *slog.Logger
	jwt     *jwtoken.TokenService
	usecase port.AuthUsecase
}

func newAuthController(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.AuthUsecase,
) *authController {
	return &authController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

// Register creates a new user
func (c *authController) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.RegisterNewUser"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userInput := &model.UserRequestData{}
		if err := decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		userDevice := model.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		tokenData, userID, err := c.usecase.RegisterNewUser(ctx, userInput, userDevice)
		switch {
		case errors.Is(err, le.ErrUserAlreadyExists):
			handleResponseError(w, r, log, http.StatusConflict, le.ErrUserAlreadyExists,
				slog.String(key.Email, userInput.Email))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToCreateUser, err)
		}

		log.Info("user and tokens created",
			slog.String(key.UserID, userID),
			slog.Any(key.AccessToken, tokenData.GetAccessToken()),
			slog.Any(key.RefreshToken, tokenData.GetRefreshToken()),
		)
		jwtoken.SendTokensToWeb(w, tokenData, http.StatusCreated)
	}
}

func (c *authController) LoginWithPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.LoginWithPassword"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userInput := &model.UserRequestData{}
		if err := decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		userDevice := model.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		tokenData, userID, err := c.usecase.LoginUser(ctx, userInput, userDevice)
		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrUserNotFound,
				slog.String(key.Email, userInput.Email))
		case errors.Is(err, le.ErrUserUnauthenticated):
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrUserUnauthenticated,
				slog.String(key.Email, userInput.Email))
		case err != nil:
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrFailedToLoginUser,
				slog.String(key.Email, userInput.Email),
				slog.String(key.Error, err.Error()))
		default:
			log.Info(
				"user logged in, tokens created",
				slog.String(key.UserID, userID),
				slog.Any(key.AccessToken, tokenData.AccessToken),
				slog.Any(key.RefreshToken, tokenData.RefreshToken),
			)
			jwtoken.SendTokensToWeb(w, tokenData, http.StatusOK)
		}
	}
}

func (c *authController) RefreshJWTokens() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.RefreshJWTokens"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		refreshToken, err := jwtoken.FindRefreshToken(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrFailedToGetRefreshToken, err)
			return
		}

		userDevice := model.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		tokenData, userID, err := c.usecase.Refresh(ctx, refreshToken, userDevice)
		switch {
		case err != nil:
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrFailedToRefreshTokens,
				slog.String(key.UserID, userID), // TODO: check if can return userID when error is here
				slog.String(key.Error, err.Error()))
		default:
			log.Info("tokens created",
				slog.Any(key.UserID, userID),
				slog.Any(key.AccessToken, tokenData.AccessToken),
				slog.Any(key.RefreshToken, tokenData.RefreshToken))
			jwtoken.SendTokensToWeb(w, tokenData, http.StatusOK)
		}
	}
}

// Logout removes user session
func (c *authController) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.Logout"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userDevice := model.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		err := c.usecase.LogoutUser(ctx, userDevice)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToLogout, err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refreshToken",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			MaxAge:   -1,
		})

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("Logged out successfully"))
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToWriteResponse, err)
			return
		}
	}
}

// GetUser get a user by ID
func (c *authController) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.GetUserData"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		user, err := c.usecase.GetUserByID(ctx)

		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound, slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "user received", user, slog.String(key.UserID, userID))
		}
	}
}

// UpdateUser updates a user by ID
func (c *authController) UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.UpdateUser"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		userInput := &model.UserRequestData{}
		if err = decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		err = c.usecase.UpdateUser(ctx, userInput)

		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound,
				slog.String(key.UserID, userID),
				slog.String(key.Email, userInput.Email))
		case errors.Is(err, le.ErrEmailAlreadyTaken):
			handleResponseError(w, r, log, http.StatusConflict, le.ErrEmailAlreadyTaken,
				slog.String(key.UserID, userID),
				slog.String(key.Email, userInput.Email))
		case errors.Is(err, le.ErrNoChangesDetected):
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrNoChangesDetected,
				slog.String(key.UserID, userID))
		default:
			handleResponseSuccess(w, r, log, "user updated",
				model.UserResponseData{ID: userID},
				slog.String(key.UserID, userID))
		}
	}
}

// DeleteUser deletes a user by ID
func (c *authController) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.DeleteUser"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		userDevice := model.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		err = c.usecase.DeleteUser(ctx, userDevice)

		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound,
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToDeleteUser, err)
		default:
			handleResponseSuccess(w, r, log, "user deleted",
				model.UserResponseData{ID: userID},
				slog.String(key.UserID, userID),
			)
		}
	}
}
