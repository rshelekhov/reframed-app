package handlers

import (
	"errors"
	"github.com/rshelekhov/reframed/src/le"
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
			handleInternalServerError(w, r, log, le.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[contextUserID].(string)

		// Decode the request body
		err = DecodeJSON(w, r, log, list)
		if err != nil {
			return
		}

		id := ksuid.New().String()
		now := time.Now().UTC()

		newList := models.List{
			ID:        id,
			Title:     list.Title,
			UserID:    userID,
			UpdatedAt: &now,
		}

		err = h.ListStorage.CreateList(r.Context(), newList)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToCreateList, err)
			return
		}

		log.Info("list created", slog.Any("list_id", id))
		responseSuccess(w, r, http.StatusCreated, "list created", models.List{ID: id})
	}
}

func (h *ListHandler) GetListByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			op = "list.handlers.GetListByID"
		)
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
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrFailedToParsePagination, err)
			return
		}

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[contextUserID].(string)

		lists, err := h.ListStorage.GetLists(r.Context(), userID, pagination)
		if errors.Is(err, le.ErrNoListsFound) {
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoListsFound)
			return
		}
		if err != nil {
			log.Error("failed to get lists", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to get lists")
			handleInternalServerError(w, r, log, le.ErrFailedToGetLists, err)
			return
		}

		log.Info(
			"lists found",
			slog.Int("count", len(lists)),
			slog.Int("limit", pagination.Limit),
			slog.Int("offset", pagination.Offset),
		)
		responseSuccess(w, r, http.StatusOK, "lists found", lists)
	}
}

func (h *ListHandler) UpdateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handlers.UpdateList"
	}
}

func (h *ListHandler) DeleteList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.handlers.DeleteList"
	}
}
