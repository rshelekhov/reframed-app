package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rshelekhov/jwtauth"
	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/port"
)

type statusHandler struct {
	logger  *slog.Logger
	jwt     *jwtauth.TokenService
	usecase port.StatusUsecase
}

func newStatusHandler(
	log *slog.Logger,
	jwt *jwtauth.TokenService,
	usecase port.StatusUsecase,
) *statusHandler {
	return &statusHandler{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

func (h *statusHandler) GetStatuses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "status.handler.GetStatuses"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		statuses, err := h.usecase.GetStatuses(ctx)

		switch {
		case errors.Is(err, le.ErrNoStatusesFound):
			handleResponseSuccess(w, r, log, "no statuses found", nil)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		}

		handleResponseSuccess(w, r, log, "statuses found", statuses)
	}
}

func (h *statusHandler) GetStatusByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "status.handler.GetStatusName"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		rawStatusID := chi.URLParam(r, key.StatusID)

		statusID, err := strconv.Atoi(rawStatusID)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrFailedToConvertStatusIDtoInt)
			return
		}
		if statusID <= 0 {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrInvalidStatusID)
			return
		}

		status, err := h.usecase.GetStatusByID(ctx, statusID)

		switch {
		case errors.Is(err, le.ErrStatusNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrStatusNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		}

		handleResponseSuccess(w, r, log, "status received", status, slog.Int(key.StatusID, status.ID))
	}
}
