package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/go-chi/chi/v5"

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

func NewTagRoutes(
	r *chi.Mux,
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.TagUsecase,
) {
	c := &tagController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtoken.Verifier(jwt))
		r.Use(jwtoken.Authenticator())

		r.Get("/user/tags", c.GetTagsByUserID())
	})
}

func (c *tagController) GetTagsByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "tag.controller.GetTagsByUserID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := jwtoken.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		tagsResp, err := c.usecase.GetTagsByUserID(ctx, userID)

		switch {
		case errors.Is(err, le.ErrNoTagsFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoTagsFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tags found", tagsResp, slog.Int(key.Count, len(tagsResp)))
		}
	}
}
