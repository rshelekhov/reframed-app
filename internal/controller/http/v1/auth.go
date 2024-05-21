package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/go-chi/chi/v5"

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

func NewAuthRoutes(
	r *chi.Mux,
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.AuthUsecase,
) {
	c := &authController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}

	// Public routes
	r.Group(func(r chi.Router) {
		r.Post("/login", c.LoginWithPassword())
		r.Post("/register", c.Register())
		// TODO: add handler for RequestResetPassword
		r.Post("/refresh-tokens", c.RefreshJWTokens())
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtoken.Verifier(jwt))
		r.Use(jwtoken.Authenticator())

		r.Post("/logout", c.Logout())

		r.Route("/user", func(r chi.Router) {
			r.Get("/", c.GetUser())
			r.Put("/", c.UpdateUser())
			r.Delete("/", c.DeleteUser())
		})
	})
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
				slog.String(key.Email, userInput.Email),
			)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToCreateUser, err)
			return
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
		if err != nil {
			if errors.Is(err, le.ErrUserNotFound) {
				handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrUserNotFound,
					slog.String(key.Email, userInput.Email),
				)
				return
			}
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrFailedToLoginUser,
				slog.String(key.Email, userInput.Email),
				slog.String(key.Error, err.Error()),
			)
		}

		log.Info(
			"user logged in, tokens created",
			slog.String(key.UserID, userID),
			slog.Any(key.AccessToken, tokenData.AccessToken),
			slog.Any(key.RefreshToken, tokenData.RefreshToken),
		)
		jwtoken.SendTokensToWeb(w, tokenData, http.StatusOK)
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
		if err != nil {
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrFailedToLoginUser,
				slog.String(key.UserID, userID), // TODO: check if can return userID when error is here
				slog.String(key.Error, err.Error()),
			)
		}

		log.Info("tokens created",
			slog.Any(key.UserID, userID),
			slog.Any(key.AccessToken, tokenData.AccessToken),
			slog.Any(key.RefreshToken, tokenData.RefreshToken))
		jwtoken.SendTokensToWeb(w, tokenData, http.StatusOK)
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
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		user, err := c.usecase.GetUserByID(ctx)

		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound, slog.String(key.UserID, userID))
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
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
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
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
				slog.String(key.Email, userInput.Email),
			)
			return
		case errors.Is(err, le.ErrEmailAlreadyTaken):
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmailAlreadyTaken,
				slog.String(key.UserID, userID),
				slog.String(key.Email, userInput.Email),
			)
			return
		case errors.Is(err, le.ErrNoChangesDetected):
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrNoChangesDetected,
				slog.String(key.UserID, userID),
			)
			return
		case errors.Is(err, le.ErrNoPasswordChangesDetected):
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrNoPasswordChangesDetected,
				slog.String(key.UserID, userID),
			)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateUser,
				slog.String(key.UserID, userID),
				slog.Any(key.Error, err),
			)
			return
		default:
			handleResponseSuccess(w, r, log, "user updated",
				model.UserResponseData{ID: userID},
				slog.String(key.UserID, userID),
			)
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
		if err != nil {
			if errors.Is(err, le.ErrUserNotFound) {
				handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound, slog.String(key.UserID, userID))
				return
			}
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		userDevice := model.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		err = c.usecase.DeleteUser(ctx, userDevice)

		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound,
				le.ErrUserNotFound,
				slog.String(key.UserID, userID),
			)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToDeleteUser, err)
			return
		default:
			handleResponseSuccess(w, r, log, "user deleted",
				model.UserResponseData{ID: userID},
				slog.String(key.UserID, userID),
			)
		}
	}
}
