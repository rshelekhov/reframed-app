package v1

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/rshelekhov/jwtauth"

	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type authHandler struct {
	logger  *slog.Logger
	jwt     *jwtauth.TokenService
	usecase port.AuthUsecase
}

func newAuthHandler(
	log *slog.Logger,
	jwt *jwtauth.TokenService,
	usecase port.AuthUsecase,
) *authHandler {
	return &authHandler{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

// Register creates a new user
func (h *authHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handler.RegisterNewUser"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userInput := &model.UserRequestData{}
		if err := decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		userDevice := model.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		tokenData, userID, err := h.usecase.RegisterNewUser(ctx, userInput, userDevice)

		switch {
		case errors.Is(err, le.ErrUserAlreadyExists):
			handleResponseError(w, r, log, http.StatusConflict, le.ErrUserAlreadyExists,
				slog.String(key.Email, userInput.Email))
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToCreateUser, err)
			return
		}

		log.Info("user and tokens created",
			slog.String(key.UserID, userID),
			slog.Any(key.AccessToken, tokenData.GetAccessToken()),
			slog.Any(key.RefreshToken, tokenData.GetRefreshToken()))
		jwtauth.SendTokensToWeb(w, tokenData, http.StatusCreated)
	}
}

func (h *authHandler) VerifyEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handler.VerifyEmail"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		verificationToken := r.URL.Query().Get(key.Token)
		if verificationToken == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmailVerificationTokenNotFoundInQuery)
			return
		}

		err := h.usecase.VerifyEmail(ctx, verificationToken)

		switch {
		case errors.Is(err, le.ErrEmailVerificationTokenExpiredWithEmailResent):
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrEmailVerificationTokenExpiredWithEmailResent)
			return
		case errors.Is(err, le.ErrEmailVerificationTokenNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrEmailVerificationTokenNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToVerifyEmail, err)
			return
		}

		handleResponseSuccess(w, r, log, "user verified", nil)
	}
}

func (h *authHandler) LoginWithPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handler.LoginWithPassword"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userInput := &model.UserRequestData{}
		if err := decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		userDevice := model.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		tokenData, userID, err := h.usecase.LoginUser(ctx, userInput, userDevice)

		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrUserNotFound,
				slog.String(key.Email, userInput.Email))
			return
		case errors.Is(err, le.ErrUserUnauthenticated):
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrUserUnauthenticated,
				slog.String(key.Email, userInput.Email))
			return
		case err != nil:
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrFailedToLoginUser,
				slog.String(key.Email, userInput.Email),
				slog.String(key.Error, err.Error()))
			return
		}

		log.Info(
			"user logged in, tokens created",
			slog.String(key.UserID, userID),
			slog.Any(key.AccessToken, tokenData.AccessToken),
			slog.Any(key.RefreshToken, tokenData.RefreshToken))
		jwtauth.SendTokensToWeb(w, tokenData, http.StatusOK)
	}
}

func (h *authHandler) RequestResetPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handler.RequestResetPassword"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userInput := &model.ResetPasswordRequestData{}
		if err := decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		err := h.usecase.RequestResetPassword(ctx, userInput.Email)

		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound,
				slog.String(key.Email, userInput.Email))
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToRequestResetPassword, err)
			return
		}

		handleResponseSuccess(w, r, log, "reset password email sent", nil)
	}
}

func (h *authHandler) ChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handler.ChangePassword"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userInput := &model.ChangePasswordRequestData{}
		if err := decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		resetPasswordToken := r.URL.Query().Get(key.Token)
		if resetPasswordToken == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrResetPasswordTokenNotFoundInQuery)
			return
		}

		err := h.usecase.ChangePassword(ctx, userInput.Password, resetPasswordToken)

		switch {
		case errors.Is(err, le.ErrResetPasswordTokenExpiredWithEmailResent):
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrResetPasswordTokenExpiredWithEmailResent)
			return
		case errors.Is(err, le.ErrUpdatedPasswordMustNotMatchTheCurrent):
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrUpdatedPasswordMustNotMatchTheCurrent)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToChangePassword, err)
			return
		}

		handleResponseSuccess(w, r, log, "password changed", nil)
	}
}

func (h *authHandler) RefreshTokens() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handler.RefreshTokens"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		refreshToken, err := jwtauth.FindRefreshToken(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrFailedToGetRefreshToken, err)
			return
		}

		userDevice := model.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		tokenData, userID, err := h.usecase.RefreshTokens(ctx, refreshToken, userDevice)
		if err != nil {
			handleResponseError(w, r, log, http.StatusUnauthorized, le.ErrFailedToRefreshTokens,
				slog.String(key.UserID, userID),
				slog.Any(key.Error, err))
			return
		}

		log.Info("tokens created",
			slog.String(key.UserID, userID),
			slog.String(key.AccessToken, tokenData.AccessToken),
			slog.String(key.RefreshToken, tokenData.RefreshToken))
		jwtauth.SendTokensToWeb(w, tokenData, http.StatusOK)
	}
}

// Logout removes user session
func (h *authHandler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handler.Logout"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		userDevice := model.UserDeviceRequestData{
			UserAgent: r.UserAgent(),
			IP:        strings.Split(r.RemoteAddr, ":")[0],
		}

		err = h.usecase.LogoutUser(ctx, userDevice)
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

		log.Info("user logged out", slog.String(key.UserID, userID))
	}
}

// GetUser get a user by ID
func (h *authHandler) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handler.GetUserData"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		user, err := h.usecase.GetUserByID(ctx)

		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound, slog.String(key.UserID, userID))
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		}

		handleResponseSuccess(w, r, log, "user received", user, slog.String(key.UserID, userID))
	}
}

// UpdateUser updates a user by ID
func (h *authHandler) UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handler.UpdateUser"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		userInput := &model.UpdateUserRequestData{}
		if err = decodeAndValidateJSON(w, r, log, userInput); err != nil {
			return
		}

		err = h.usecase.UpdateUser(ctx, userInput)

		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound,
				slog.String(key.UserID, userID),
				slog.String(key.Email, userInput.Email))
			return
		case errors.Is(err, le.ErrEmailAlreadyTaken):
			handleResponseError(w, r, log, http.StatusConflict, le.ErrEmailAlreadyTaken,
				slog.String(key.UserID, userID),
				slog.String(key.Email, userInput.Email))
			return
		case err != nil:
			errStr := fmt.Sprint(err)
			if strings.Contains(errStr, "bad request") {
				handleResponseError(w, r, log, http.StatusBadRequest, le.LocalError(errStr),
					slog.String(key.UserID, userID))
				return
			} else {
				handleInternalServerError(w, r, log, le.ErrFailedToUpdateUser, err)
				return
			}
		}

		handleResponseSuccess(w, r, log, "user updated",
			model.UserResponseData{ID: userID},
			slog.String(key.UserID, userID))
	}
}

// DeleteUser deletes a user by ID
func (h *authHandler) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "auth.handler.DeleteUser"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		err = h.usecase.DeleteUser(ctx, userID)

		switch {
		case errors.Is(err, le.ErrUserNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrUserNotFound,
				slog.String(key.UserID, userID))
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToDeleteUser, err)
			return
		}

		handleResponseSuccess(w, r, log, "user deleted",
			model.UserResponseData{ID: userID},
			slog.String(key.UserID, userID))
	}
}
