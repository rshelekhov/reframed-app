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

type listController struct {
	logger  *slog.Logger
	jwt     *jwtoken.TokenService
	usecase port.ListUsecase
}

func newListController(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.ListUsecase,
) *listController {
	return &listController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

func (c *listController) CreateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.controller.CreateList"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listInput := &model.ListRequestData{}
		if err = decodeAndValidateJSON(w, r, log, listInput); err != nil {
			return
		}

		listInput.UserID = userID

		list, err := c.usecase.CreateList(ctx, listInput)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToCreateList, err)
			return
		}

		handleResponseCreated(w, r, log, "list created", list, slog.String(key.ListID, list.ID))
	}
}

func (c *listController) GetDefaultList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.controller.GetDefaultList"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listID, err := c.usecase.GetDefaultListID(ctx, userID)
		switch {
		case errors.Is(err, le.ErrDefaultListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrDefaultListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		}

		listInput := model.ListRequestData{
			ID:     listID,
			UserID: userID,
		}

		listResp, err := c.usecase.GetListByID(ctx, listInput)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		}

		handleResponseSuccess(w, r, log, "default list received", listResp, slog.String(key.ListID, listID))
	}
}

func (c *listController) GetListByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.controller.GetListByID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listID := chi.URLParam(r, "list_id")

		listInput := model.ListRequestData{
			ID:     listID,
			UserID: userID,
		}

		listResp, err := c.usecase.GetListByID(ctx, listInput)

		switch {
		case errors.Is(err, le.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "list received", listResp, slog.String(key.ListID, listID))
		}
	}
}

func (c *listController) GetListsByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.controller.GetListsByUserID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listsResp, err := c.usecase.GetListsByUserID(ctx, userID)

		switch {
		case errors.Is(err, le.ErrNoListsFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoListsFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetLists, err)
			return
		default:
			handleResponseSuccess(w, r, log, "lists found", listsResp,
				slog.Int(key.Count, len(listsResp)),
			)
		}
	}
}

func (c *listController) UpdateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.controller.UpdateList"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listID := chi.URLParam(r, key.ListID)

		listInput := &model.ListRequestData{}
		if err = decodeAndValidateJSON(w, r, log, listInput); err != nil {
			return
		}

		listInput.ID = listID
		listInput.UserID = userID

		listResponse, err := c.usecase.UpdateList(ctx, listInput)

		switch {
		case errors.Is(err, le.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateList, err)
			return
		default:
			handleResponseSuccess(w, r, log, "list updated", listResponse, slog.String(key.ListID, listResponse.ID))
		}
	}
}

func (c *listController) DeleteList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.controller.DeleteList"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listID := chi.URLParam(r, key.ListID)

		listInput := model.ListRequestData{
			ID:     listID,
			UserID: userID,
		}

		err = c.usecase.DeleteList(ctx, listInput)

		switch {
		case errors.Is(err, le.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrListNotFound)
			return
		case errors.Is(err, le.ErrCannotDeleteDefaultList):
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrCannotDeleteDefaultList)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToDeleteList, err)
			return
		default:
			handleResponseSuccess(w, r, log, "list deleted", listID, slog.String(key.ListID, listID))
		}
	}
}
