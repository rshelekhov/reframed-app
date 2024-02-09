package handlers

import (
	"errors"
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/logger"
	"github.com/rshelekhov/reframed/src/server/middleware/jwtoken"
	"github.com/rshelekhov/reframed/src/storage"
	"log/slog"
	"net/http"
)

type TagHandler struct {
	logger     logger.Interface
	tokenAuth  *jwtoken.JWTAuth
	tagStorage storage.TagStorage
}

func NewTagHandler(
	log logger.Interface,
	tokenAuth *jwtoken.JWTAuth,
	tagStorage storage.TagStorage,
) *TagHandler {
	return &TagHandler{
		logger:     log,
		tokenAuth:  tokenAuth,
		tagStorage: tagStorage,
	}
}

func (h *TagHandler) GetTagsByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "tag.handlers.GetTagsByUserID"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		tags, err := h.tagStorage.GetTagsByUserID(r.Context(), userID)
		switch {
		case errors.Is(err, c.ErrNoTagsFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTagsFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tags found", tags, slog.Int(c.CountKey, len(tags)))
		}
	}
}
