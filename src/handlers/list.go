package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/logger"
	"github.com/rshelekhov/reframed/src/models"
	"github.com/rshelekhov/reframed/src/server/middleware/jwtoken"
	"github.com/rshelekhov/reframed/src/storage"
	"github.com/segmentio/ksuid"
	"log/slog"
	"net/http"
)

type ListHandler struct {
	logger         logger.Interface
	tokenAuth      *jwtoken.JWTAuth
	listStorage    storage.ListStorage
	headingStorage storage.HeadingStorage
}

func NewListHandler(
	log logger.Interface,
	tokenAuth *jwtoken.JWTAuth,
	listStorage storage.ListStorage,
	headingStorage storage.HeadingStorage,
) *ListHandler {
	return &ListHandler{
		logger:         log,
		tokenAuth:      tokenAuth,
		listStorage:    listStorage,
		headingStorage: headingStorage,
	}
}

func (h *ListHandler) CreateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handlers.CreateList"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		list := &models.List{}

		if err = DecodeAndValidateJSON(w, r, log, list); err != nil {
			return
		}

		newList := models.List{
			ID:     ksuid.New().String(),
			Title:  list.Title,
			UserID: userID,
		}

		if err = h.listStorage.CreateList(r.Context(), newList); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToCreateList, err)
			return
		}

		if err = h.headingStorage.CreateHeading(
			r.Context(),
			models.Heading{
				ID:     ksuid.New().String(),
				Title:  c.DefaultHeadingTitle,
				ListID: newList.ID,
				UserID: userID,
			},
			true,
		); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToCreateList, err)
			return
		}

		handleResponseCreated(
			w, r, log, "list created",
			models.List{ID: newList.ID},
			slog.String(c.ListIDKey, newList.ID),
		)
	}
}

func (h *ListHandler) GetListByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			op = "list.handlers.GetListByID"
		)

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID := chi.URLParam(r, c.ListIDKey)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryListID)
			return
		}

		list, err := h.listStorage.GetListByID(r.Context(), listID, userID)
		switch {
		case errors.Is(err, c.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "list received", list, slog.String(c.ListIDKey, listID))
		}
	}
}

func (h *ListHandler) GetListsByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			op = "list.handlers.GetListsByUserID"
		)

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		lists, err := h.listStorage.GetListsByUserID(r.Context(), userID)
		switch {
		case errors.Is(err, c.ErrNoListsFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoListsFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetLists, err)
			return
		default:
			handleResponseSuccess(w, r, log, "lists found", lists,
				slog.Int(c.CountKey, len(lists)),
			)
		}
	}
}

func (h *ListHandler) UpdateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handlers.UpdateList"

		log := logger.LogWithRequest(h.logger, op, r)
		list := &models.UpdateList{}

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID := chi.URLParam(r, c.ListIDKey)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryListID)
			return
		}

		if err = DecodeAndValidateJSON(w, r, log, list); err != nil {
			return
		}

		updatedList := models.List{
			ID:     listID,
			Title:  list.Title,
			UserID: userID,
		}

		err = h.listStorage.UpdateList(r.Context(), updatedList)
		switch {
		case errors.Is(err, c.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToUpdateList, err)
			return
		default:
			handleResponseSuccess(w, r, log, "list updated", updatedList, slog.String(c.ListIDKey, listID))
		}
	}
}

func (h *ListHandler) DeleteList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handlers.DeleteList"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID := chi.URLParam(r, c.ListIDKey)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryListID)
			return
		}

		err = h.listStorage.DeleteList(r.Context(), listID, userID)
		switch {
		case errors.Is(err, c.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToDeleteList, err)
			return
		default:
			handleResponseSuccess(w, r, log, "list deleted", models.List{ID: listID}, slog.String(c.ListIDKey, listID))
		}
	}
}
