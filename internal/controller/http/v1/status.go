package v1

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/port"
	"log/slog"
	"net/http"
	"strconv"
)

type statusController struct {
	logger  *slog.Logger
	jwt     *jwtoken.TokenService
	usecase port.StatusUsecase
}

func newStatusController(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.StatusUsecase,
) *statusController {
	return &statusController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

func (c *statusController) GetStatuses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "status.controller.GetStatuses"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		statuses, err := c.usecase.GetStatuses(ctx)

		switch {
		case errors.Is(err, le.ErrNoStatusesFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoStatusesFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "statuses received", statuses, slog.Int(key.Count, len(statuses)))
			return
		}
	}
}

func (c *statusController) GetStatusByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "status.controller.GetStatusName"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		rawStatusID := chi.URLParam(r, key.StatusID)
		if rawStatusID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryStatusID)
			return
		}

		statusID, err := strconv.Atoi(rawStatusID)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrFailedToConvertStatusIDtoInt)
			return
		}
		if statusID <= 0 {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrInvalidStatusID)
			return
		}

		status, err := c.usecase.GetStatusByID(ctx, statusID)

		switch {
		case errors.Is(err, le.ErrStatusNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrStatusNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "status received", status, slog.Int(key.StatusID, status.ID))
		}
	}
}
