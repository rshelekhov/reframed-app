package handlers

import (
	"errors"
	"fmt"
	"github.com/rshelekhov/reframed/internal/http-server/middleware/auth"
	"github.com/rshelekhov/reframed/internal/logger"
	"github.com/rshelekhov/reframed/internal/models"
	"github.com/rshelekhov/reframed/internal/storage"
	"github.com/segmentio/ksuid"
	"log/slog"
	"net/http"
	"time"
)

type ListHandler struct {
	Logger      logger.Interface
	TokenAuth   *auth.JWTAuth
	ListStorage storage.ListStorage
}

func NewListHandler(
	log logger.Interface,
	tokenAuth *auth.JWTAuth,
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

		// Decode the request body
		err := DecodeJSON(w, r, log, list)
		if err != nil {
			return
		}

		id := ksuid.New().String()
		now := time.Now().UTC()

		newList := models.List{
			ID:        id,
			Title:     list.Title,
			UserID:    list.UserID,
			UpdatedAt: &now,
		}

		err = h.ListStorage.CreateList(r.Context(), newList)
		if err != nil {
			log.Error("failed to create list", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to create list")
			return
		}

		log.Info("list created", slog.Any("list_id", id))
		responseSuccess(w, r, http.StatusCreated, "list created", models.List{ID: id})
	}
}

func (h *ListHandler) GetListByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			op  = "list.handlers.GetListByID"
			key = "listID"
		)
	}
}

func (h *ListHandler) GetListsByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const (
			op  = "list.handlers.GetLists"
			key = "user_id"
		)

		log := logger.LogWithRequest(h.Logger, op, r)

		pagination, err := ParseLimitAndOffset(r)
		if err != nil {
			log.Error(ErrFailedToParsePagination.Error(), logger.Err(err))
			responseError(w, r, http.StatusBadRequest, ErrFailedToParsePagination.Error())
			return
		}

		// TODO: implement JWT auth
		userID, statusCode, err := GetID(r, log, key)
		if err != nil {
			responseError(w, r, statusCode, err.Error())
			return
		}

		lists, err := h.ListStorage.GetLists(r.Context(), userID, pagination)
		if errors.Is(err, storage.ErrNoListsFound) {
			log.Error(fmt.Sprintf("%v", storage.ErrNoListsFound))
			responseError(w, r, http.StatusNotFound, fmt.Sprintf("%v", storage.ErrNoListsFound))
			return
		}
		if err != nil {
			log.Error("failed to get lists", logger.Err(err))
			responseError(w, r, http.StatusInternalServerError, "failed to get lists")
			return
		}

		log.Info(
			"users found",
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
