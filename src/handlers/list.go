package handlers

import (
	"errors"
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/logger"
	"github.com/rshelekhov/reframed/src/models"
	"github.com/rshelekhov/reframed/src/server/middleware/jwtoken"
	"github.com/rshelekhov/reframed/src/storage"
	"github.com/segmentio/ksuid"
	"log/slog"
	"net/http"
	"time"
)

type ListHandler struct {
	Logger      logger.Interface
	TokenAuth   *jwtoken.JWTAuth
	ListStorage storage.ListStorage
}

func NewListHandler(
	log logger.Interface,
	tokenAuth *jwtoken.JWTAuth,
	listStorage storage.ListStorage,
) *ListHandler {
	return &ListHandler{
		Logger:      log,
		TokenAuth:   tokenAuth,
		ListStorage: listStorage,
	}
}

func (h *ListHandler) CreateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handlers.CreateList"

		log := logger.LogWithRequest(h.Logger, op, r)

		list := &models.List{}

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		err = DecodeAndValidateJSON(w, r, log, list)
		if err != nil {
			return
		}

		now := time.Now().UTC()

		newList := models.List{
			ID:        ksuid.New().String(),
			Title:     list.Title,
			UserID:    userID,
			UpdatedAt: &now,
		}

		err = h.ListStorage.CreateList(r.Context(), newList)
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToCreateList, err)
			return
		}

		handleResponseSuccess(
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

		log := logger.LogWithRequest(h.Logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID, err := GetIDFromQuery(w, r, log, c.ListIDKey)
		if err != nil {
			return
		}

		list, err := h.ListStorage.GetListByID(r.Context(), listID, userID)
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

		log := logger.LogWithRequest(h.Logger, op, r)

		pagination, err := ParseLimitAndOffset(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrFailedToParsePagination, err)
			return
		}

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		lists, err := h.ListStorage.GetLists(r.Context(), userID, pagination)
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
				slog.Int(c.LimitKey, pagination.Limit),
				slog.Int(c.OffsetKey, pagination.Offset),
			)
		}
	}
}

func (h *ListHandler) UpdateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handlers.UpdateList"

		log := logger.LogWithRequest(h.Logger, op, r)
		list := &models.UpdateList{}

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID, err := GetIDFromQuery(w, r, log, c.ListIDKey)
		if err != nil {
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

		err = h.ListStorage.UpdateList(r.Context(), updatedList)
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

		log := logger.LogWithRequest(h.Logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID, err := GetIDFromQuery(w, r, log, c.ListIDKey)
		if err != nil {
			return
		}

		err = h.ListStorage.DeleteList(r.Context(), listID, userID)
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
