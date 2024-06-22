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
	logger  *slog.Logger
	jwt     *jwtoken.TokenService
	usecase port.HeadingUsecase
}

func newHeadingController(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.HeadingUsecase,
) *headingController {
	return &headingController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

func (c *headingController) CreateHeading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.CreateHeading"

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

		listID := chi.URLParam(r, key.ListID)

		headingInput := &model.HeadingRequestData{}
		if err = decodeAndValidateJSON(w, r, log, headingInput); err != nil {
			return
		}

		headingInput.ListID = listID
		headingInput.UserID = userID

		headingResponse, err := c.usecase.CreateHeading(ctx, headingInput)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToCreateHeading, err)
		}

		handleResponseCreated(w, r, log, "heading created", headingResponse,
			slog.String(key.HeadingID, headingResponse.ID))
	}
}

func (c *headingController) GetHeadingByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.controller.GetHeadingByID"

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

		headingID := chi.URLParam(r, key.HeadingID)

		headingInput := model.HeadingRequestData{
			ID:     headingID,
			UserID: userID,
		}

		headingResp, err := c.usecase.GetHeadingByID(ctx, headingInput)
		switch {
		case errors.Is(err, le.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
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

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		listID := chi.URLParam(r, key.ListID)

		headingsInput := model.HeadingRequestData{
			ListID: listID,
			UserID: userID,
		}

		headingsResp, err := c.usecase.GetHeadingsByListID(ctx, headingsInput)
		switch {
		case errors.Is(err, le.ErrNoHeadingsFound):
			handleResponseSuccess(w, r, log, "no headings found", nil,
				slog.Int(key.Count, len(headingsResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetHeadingsByListID, err)
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

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		headingID := chi.URLParam(r, key.HeadingID)

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
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateHeading, err)
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

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		headingID := chi.URLParam(r, key.HeadingID)

		otherListID := r.URL.Query().Get(key.ListID)
		if otherListID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryListID)
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
		case errors.Is(err, le.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrListNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToMoveHeading, err)
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

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		headingID := chi.URLParam(r, key.HeadingID)

		headingInput := model.HeadingRequestData{
			ID:     headingID,
			UserID: userID,
		}

		err = c.usecase.DeleteHeading(ctx, headingInput)
		switch {
		case errors.Is(err, le.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToDeleteHeading, err)
		default:
			handleResponseSuccess(w, r, log, "heading deleted", headingID, slog.String(key.HeadingID, headingID))
		}
	}
}
