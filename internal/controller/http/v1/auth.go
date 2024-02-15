package v1

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rshelekhov/reframed/internal/domain"
	"github.com/rshelekhov/reframed/pkg/constants/key"
	"github.com/rshelekhov/reframed/pkg/httpserver/middleware/jwtoken"
	"github.com/rshelekhov/reframed/pkg/logger"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type authController struct {
	logger  logger.Interface
	jwt     *jwtoken.TokenService
	usecase domain.AuthUsecase
}

func NewAuthRoutes(
	r *chi.Mux,
	log logger.Interface,
	jwt *jwtoken.TokenService,
	usecase domain.AuthUsecase,
) {
	c := &authController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}

	// Public routes
	r.Group(func(r chi.Router) {
		r.Post("/login", c.LoginWithPassword())
		r.Post("/register", c.CreateUser())
		// TODO: add handler for RequestResetPassword
		r.Post("/refresh-tokens", c.RefreshJWTokens())
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtoken.Verifier(jwt))
		r.Use(jwtoken.Authenticator())

		r.Post("/logout", c.Logout())

		r.Route("/user/", func(r chi.Router) {
			r.Get("/", c.GetUserProfile())
			r.Put("/", c.UpdateUser())
			r.Delete("/", c.DeleteUser())
		})
	})
}

// CreateUser creates a new user
func (c *authController) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.CreateUser"

		ctx := r.Context()

		log := logger.LogWithRequest(c.logger, op, r)

		userInput := &domain.UserRequestData{}
		if err := decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		// Create the user
		userID, err := c.usecase.CreateUser(ctx, c.jwt, userInput)
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrUserAlreadyExists, slog.String(key.Email, userInput.Email))
			return
		}
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToCreateUser, err)
			return
		}

		// Create session
		userDevice := domain.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		tokenData, err := c.usecase.CreateUserSession(ctx, c.jwt, userID, userDevice)
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToCreateSession, err)
			return
		}

		log.Info("user and tokens created",
			slog.String(key.UserID, userID),
			slog.Any(key.AccessToken, tokenData.AccessToken),
			slog.Any(key.RefreshToken, tokenData.RefreshToken),
		)
		jwtoken.SendTokensToWeb(w, tokenData, http.StatusCreated)
	}
}

func (c *authController) LoginWithPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.LoginWithPassword"

		ctx := r.Context()

		log := logger.LogWithRequest(c.logger, op, r)

		userInput := &domain.UserRequestData{}
		if err := decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		userID, err := c.usecase.LoginUser(ctx, c.jwt, userInput)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusUnauthorized, domain.ErrUserNotFound, slog.String(key.Email, userInput.Email))
			return
		case errors.Is(err, domain.ErrUserHasNoPassword):
			handleResponseError(w, r, log, http.StatusUnauthorized, domain.ErrUserHasNoPassword, slog.String(key.Email, userInput.Email))
			return
		case errors.Is(err, domain.ErrInvalidCredentials):
			handleResponseError(w, r, log, http.StatusUnauthorized, domain.ErrInvalidCredentials, slog.String(key.Email, userInput.Email))
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToLogin, err)
			return
		}

		// Create session
		userDevice := domain.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		tokenData, err := c.usecase.CreateUserSession(ctx, c.jwt, userID, userDevice)
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToCreateSession, err)
			return
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
			handleResponseError(w, r, log, http.StatusUnauthorized, domain.ErrFailedToGetRefreshToken, err)
			return
		}

		userDevice := domain.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
		}

		session, err := c.usecase.CheckSessionAndDevice(ctx, refreshToken, userDevice)
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			handleResponseError(w, r, log, http.StatusUnauthorized, domain.ErrSessionNotFound, err, slog.String(key.RefreshToken, refreshToken))
			return
		case errors.Is(err, domain.ErrSessionExpired):
			handleResponseError(w, r, log, http.StatusUnauthorized, domain.ErrSessionExpired, err, slog.String(key.RefreshToken, refreshToken))
			return
		case errors.Is(err, domain.ErrUserDeviceNotFound):
			handleResponseError(w, r, log, http.StatusUnauthorized, domain.ErrUserDeviceNotFound, err, slog.String(key.RefreshToken, refreshToken))
		}
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToRefreshTokens, err)
			return
		}

		tokenData, err := c.usecase.CreateUserSession(ctx, c.jwt, session.UserID, userDevice)
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToCreateSession, err)
			return
		}

		log.Info("tokens created",
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
		userID := jwtoken.GetUserID(ctx).(string)

		userDevice := domain.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
		}

		err := c.usecase.LogoutUser(ctx, userID, userDevice)
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToLogout, err)
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
			handleInternalServerError(w, r, log, domain.ErrFailedToWriteResponse, err)
			return
		}
	}
}

// GetUserProfile get a user by ID
func (c *authController) GetUserProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.controller.GetUserData"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		user, err := c.usecase.GetUserByID(ctx, userID)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrUserNotFound, slog.String(key.UserID, userID))
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		userInput := &domain.UserRequestData{}
		if err := decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		err := c.usecase.UpdateUser(ctx, c.jwt, userInput, userID)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrUserNotFound,
				slog.String(key.UserID, userID),
				slog.String(key.Email, userInput.Email),
			)
			return
		case errors.Is(err, domain.ErrEmailAlreadyTaken):
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmailAlreadyTaken,
				slog.String(key.UserID, userID),
				slog.String(key.Email, userInput.Email),
			)
			return
		case errors.Is(err, domain.ErrNoChangesDetected):
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrNoChangesDetected,
				slog.String(key.UserID, userID),
			)
			return
		case errors.Is(err, domain.ErrNoPasswordChangesDetected):
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrNoPasswordChangesDetected,
				slog.String(key.UserID, userID),
			)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToUpdateUser,
				slog.String(key.UserID, userID),
				slog.Any(key.Error, err),
			)
			return
		default:
			handleResponseSuccess(w, r, log, "user updated",
				domain.UserResponseData{ID: userID},
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
		userID := jwtoken.GetUserID(ctx).(string)

		userDevice := domain.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
		}

		err := c.usecase.DeleteUser(r.Context(), userID, userDevice)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound,
				domain.ErrUserNotFound,
				slog.String(key.UserID, userID),
			)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToDeleteUser, err)
			return
		default:
			handleResponseSuccess(w, r, log, "user deleted",
				domain.UserResponseData{ID: userID},
				slog.String(key.UserID, userID),
			)
		}
	}
}
