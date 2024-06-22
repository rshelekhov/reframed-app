package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/port"
)

type tagController struct {
	logger  *slog.Logger
	jwt     *jwtoken.TokenService
	usecase port.TagUsecase
}

func newTagController(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.TagUsecase,
) *tagController {
	return &tagController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

func (c *tagController) GetTagsByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "tag.controller.GetTagsByUserID"

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

		tagsResp, err := c.usecase.GetTagsByUserID(ctx, userID)

		switch {
		case errors.Is(err, le.ErrNoTagsFound):
			handleResponseSuccess(w, r, log, "no tags found", nil,
				slog.Int(key.Count, len(tagsResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "tags found", tagsResp, slog.Int(key.Count, len(tagsResp)))
		}
	}
}
