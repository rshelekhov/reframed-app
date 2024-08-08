package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/port"
)

type tagHandler struct {
	logger  *slog.Logger
	jwt     *jwtoken.TokenService
	usecase port.TagUsecase
}

func newTagHandler(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.TagUsecase,
) *tagHandler {
	return &tagHandler{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

func (h *tagHandler) GetTagsByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "tag.handler.GetTagsByUserID"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		tagsResp, err := h.usecase.GetTagsByUserID(ctx, userID)

		switch {
		case errors.Is(err, le.ErrNoTagsFound):
			handleResponseSuccess(w, r, log, "no tags found", nil)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "tags found", tagsResp)
		}
	}
}
