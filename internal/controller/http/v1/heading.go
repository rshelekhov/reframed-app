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
)

type headingController struct {
	logger  logger.Interface
	jwt     *jwtoken.TokenService
	usecase domain.HeadingUsecase
}

func NewHeadingRoutes(
	r *chi.Mux,
	log logger.Interface,
	jwt *jwtoken.TokenService,
	usecase domain.HeadingUsecase,
) {
	c := &headingController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}

	r.Group(func(r chi.Router) {
		r.Use(jwtoken.Verifier(jwt))
		r.Use(jwtoken.Authenticator())

		r.Route("/user/lists/{list_id}/headings", func(r chi.Router) {
			r.Post("/", c.CreateHeading())
			r.Get("/", c.GetHeadingsByListID())

			r.Route("/{heading_id}", func(r chi.Router) {
				r.Get("/", c.GetHeadingByID())
				r.Put("/", c.UpdateHeading())
				r.Put("/move/", c.MoveHeadingToAnotherList())
				r.Delete("/", c.DeleteHeading())

			})
		})
	})
}

func (c *headingController) CreateHeading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.CreateHeading"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryListID)
			return
		}

		headingInput := &domain.HeadingRequestData{}
		if err := decodeAndValidateJSON(w, r, log, headingInput); err != nil {
			return
		}

		headingInput.ListID = listID
		headingInput.UserID = userID

		headingID, err := c.usecase.CreateHeading(ctx, headingInput)
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToCreateHeading, err)
			return
		}

		handleResponseCreated(
			w, r, log, "heading created",
			domain.HeadingResponseData{ID: headingID},
			slog.String(key.HeadingID, headingID),
		)
	}
}

func (c *headingController) GetHeadingByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.GetHeadingByID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		headingID := chi.URLParam(r, key.HeadingID)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryHeadingID)
			return
		}

		headingInput := domain.HeadingRequestData{
			ID:     headingID,
			UserID: userID,
		}

		headingResp, err := c.usecase.GetHeadingByID(ctx, headingInput)
		switch {
		case errors.Is(err, domain.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading received", headingResp, slog.String(key.HeadingID, headingID))
		}
	}
}

func (c *headingController) GetHeadingsByListID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.GetHeadingsByListID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryListID)
			return
		}

		headingsInput := domain.HeadingRequestData{
			ListID: listID,
			UserID: userID,
		}

		headingsResp, err := c.usecase.GetHeadingsByListID(ctx, headingsInput)
		switch {
		case errors.Is(err, domain.ErrNoHeadingsFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoHeadingsFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetHeadingsByListID, err)
			return
		default:
			handleResponseSuccess(w, r, log, "headings found", headingsResp, slog.Int(key.Count, len(headingsResp)))
		}
	}
}

func (c *headingController) UpdateHeading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.UpdateHeading"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		headingID := chi.URLParam(r, key.HeadingID)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryHeadingID)
			return
		}

		headingInput := &domain.HeadingRequestData{}
		if err := decodeAndValidateJSON(w, r, log, headingInput); err != nil {
			return
		}

		headingInput.ID = headingID
		headingInput.UserID = userID

		err := c.usecase.UpdateHeading(ctx, headingInput)
		switch {
		case errors.Is(err, domain.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToUpdateHeading, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading updated", headingID, slog.String(key.HeadingID, headingID))
		}
	}
}

func (c *headingController) MoveHeadingToAnotherList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.MoveTaskToAnotherList"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		headingID := chi.URLParam(r, key.HeadingID)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryHeadingID)
			return
		}

		otherListID := r.URL.Query().Get(key.ListID)
		if otherListID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryListID)
			return
		}

		headingInput := domain.HeadingRequestData{
			ID:     headingID,
			ListID: otherListID,
			UserID: userID,
		}

		err := c.usecase.MoveHeadingToAnotherList(ctx, headingInput)
		switch {
		case errors.Is(err, domain.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToMoveHeading, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading moved to another list", headingID, slog.String(key.HeadingID, headingID))
		}
	}
}

func (c *headingController) DeleteHeading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.DeleteHeading"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		headingID := chi.URLParam(r, key.HeadingID)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryHeadingID)
			return
		}

		headingInput := domain.HeadingRequestData{
			ID:     headingID,
			UserID: userID,
		}

		err := c.usecase.DeleteHeading(ctx, headingInput)
		switch {
		case errors.Is(err, domain.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToDeleteHeading, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading deleted", domain.ListResponseData{ID: headingID}, slog.String(key.HeadingID, headingID))
		}
	}
}
