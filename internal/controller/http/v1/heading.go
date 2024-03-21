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
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type headingController struct {
	logger  logger.Interface
	jwt     *jwtoken.TokenService
	usecase port.HeadingUsecase
}

func NewHeadingRoutes(
	r *chi.Mux,
	log logger.Interface,
	jwt *jwtoken.TokenService,
	usecase port.HeadingUsecase,
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

		userID, err := jwtoken.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryListID)
			return
		}

		headingInput := &model.HeadingRequestData{}
		if err = decodeAndValidateJSON(w, r, log, headingInput); err != nil {
			return
		}

		headingInput.ListID = listID
		headingInput.UserID = userID

		headingResponse, err := c.usecase.CreateHeading(ctx, headingInput)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToCreateHeading, err)
			return
		}

		handleResponseCreated(
			w, r, log, "heading created",
			headingResponse,
			slog.String(key.HeadingID, headingResponse.ID),
		)
	}
}

func (c *headingController) GetHeadingByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.GetHeadingByID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := jwtoken.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		headingID := chi.URLParam(r, key.HeadingID)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryHeadingID)
			return
		}

		headingInput := model.HeadingRequestData{
			ID:     headingID,
			UserID: userID,
		}

		headingResp, err := c.usecase.GetHeadingByID(ctx, headingInput)

		switch {
		case errors.Is(err, le.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
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

		userID, err := jwtoken.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryListID)
			return
		}

		headingsInput := model.HeadingRequestData{
			ListID: listID,
			UserID: userID,
		}

		headingsResp, err := c.usecase.GetHeadingsByListID(ctx, headingsInput)

		switch {
		case errors.Is(err, le.ErrNoHeadingsFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoHeadingsFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetHeadingsByListID, err)
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

		userID, err := jwtoken.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		headingID := chi.URLParam(r, key.HeadingID)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryHeadingID)
			return
		}

		headingInput := &model.HeadingRequestData{}
		if err = decodeAndValidateJSON(w, r, log, headingInput); err != nil {
			return
		}

		headingInput.ID = headingID
		headingInput.UserID = userID

		headingResponse, err := c.usecase.UpdateHeading(ctx, headingInput)

		switch {
		case errors.Is(err, le.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateHeading, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading updated", headingResponse, slog.String(key.HeadingID, headingResponse.ID))
		}
	}
}

func (c *headingController) MoveHeadingToAnotherList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.MoveTaskToAnotherList"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := jwtoken.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		headingID := chi.URLParam(r, key.HeadingID)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryHeadingID)
			return
		}

		otherListID := r.URL.Query().Get(key.ListID)
		if otherListID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryListID)
			return
		}

		headingInput := model.HeadingRequestData{
			ID:     headingID,
			ListID: otherListID,
			UserID: userID,
		}

		headingResponse, err := c.usecase.MoveHeadingToAnotherList(ctx, headingInput)

		switch {
		case errors.Is(err, le.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToMoveHeading, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading moved to another list", headingResponse, slog.String(key.HeadingID, headingResponse.ID))
		}
	}
}

func (c *headingController) DeleteHeading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.DeleteHeading"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := jwtoken.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		headingID := chi.URLParam(r, key.HeadingID)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryHeadingID)
			return
		}

		headingInput := model.HeadingRequestData{
			ID:     headingID,
			UserID: userID,
		}

		err = c.usecase.DeleteHeading(ctx, headingInput)

		switch {
		case errors.Is(err, le.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToDeleteHeading, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading deleted", headingID, slog.String(key.HeadingID, headingID))
		}
	}
}
