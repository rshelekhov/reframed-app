package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/go-chi/chi/v5"

	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type listHandler struct {
	logger  *slog.Logger
	jwt     *jwtoken.TokenService
	usecase port.ListUsecase
}

func newListHandler(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.ListUsecase,
) *listHandler {
	return &listHandler{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

func (h *listHandler) CreateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handler.CreateList"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		listInput := &model.ListRequestData{}
		if err = decodeAndValidateJSON(w, r, log, listInput); err != nil {
			return
		}

		listInput.UserID = userID

		list, err := h.usecase.CreateList(ctx, listInput)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToCreateList, err)
			return
		}

		handleResponseCreated(w, r, log, "list created", list, slog.String(key.ListID, list.ID))
	}
}

func (h *listHandler) GetDefaultList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handler.GetDefaultList"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		listID, err := h.usecase.GetDefaultListID(ctx, userID)

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

		listResp, err := h.usecase.GetListByID(ctx, listInput)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		}

		handleResponseSuccess(w, r, log, "default list received", listResp, slog.String(key.ListID, listID))
	}
}

func (h *listHandler) GetListByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handler.GetListByID"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		listID := chi.URLParam(r, "list_id")

		listInput := model.ListRequestData{
			ID:     listID,
			UserID: userID,
		}

		listResp, err := h.usecase.GetListByID(ctx, listInput)

		switch {
		case errors.Is(err, le.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		}

		handleResponseSuccess(w, r, log, "list received", listResp, slog.String(key.ListID, listID))
	}
}

func (h *listHandler) GetListsByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handler.GetListsByUserID"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		listsResp, err := h.usecase.GetListsByUserID(ctx, userID)

		switch {
		case errors.Is(err, le.ErrNoListsFound):
			handleResponseSuccess(w, r, log, "no lists found", nil)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetLists, err)
			return
		}

		handleResponseSuccess(w, r, log, "lists found", listsResp)
	}
}

func (h *listHandler) UpdateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handler.UpdateList"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		listID := chi.URLParam(r, key.ListID)

		listInput := &model.ListRequestData{}
		if err = decodeAndValidateJSON(w, r, log, listInput); err != nil {
			return
		}

		listInput.ID = listID
		listInput.UserID = userID

		listResponse, err := h.usecase.UpdateList(ctx, listInput)

		switch {
		case errors.Is(err, le.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateList, err)
			return
		}

		handleResponseSuccess(w, r, log, "list updated", listResponse, slog.String(key.ListID, listResponse.ID))
	}
}

func (h *listHandler) DeleteList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handler.DeleteList"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		listID := chi.URLParam(r, key.ListID)

		listInput := model.ListRequestData{
			ID:     listID,
			UserID: userID,
		}

		err = h.usecase.DeleteList(ctx, listInput)

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
		}

		handleResponseSuccess(w, r, log, "list deleted", listID, slog.String(key.ListID, listID))
	}
}
