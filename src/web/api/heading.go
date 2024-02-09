package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/logger"
	"github.com/rshelekhov/reframed/src/models"
	"github.com/rshelekhov/reframed/src/server/middleware/jwtoken/service"
	"github.com/rshelekhov/reframed/src/storage"
	"github.com/segmentio/ksuid"
	"log/slog"
	"net/http"
)

type HeadingHandler struct {
	logger         logger.Interface
	tokenAuth      *service.JWTokenService
	headingStorage storage.HeadingStorage
}

func NewHeadingHandler(
	log logger.Interface,
	tokenAuth *service.JWTokenService,
	headingStorage storage.HeadingStorage,
) *HeadingHandler {
	return &HeadingHandler{
		logger:         log,
		tokenAuth:      tokenAuth,
		headingStorage: headingStorage,
	}
}

func (h *HeadingHandler) CreateHeading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.handlers.CreateHeading"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
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

		heading := &models.Heading{}

		if err = DecodeAndValidateJSON(w, r, log, heading); err != nil {
			return
		}

		newHeading := models.Heading{
			ID:     ksuid.New().String(),
			Title:  heading.Title,
			ListID: listID,
			UserID: userID,
		}

		if err = h.headingStorage.CreateHeading(r.Context(), newHeading, false); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToCreateHeading, err)
			return
		}

		handleResponseCreated(
			w, r, log, "heading created",
			models.Heading{ID: newHeading.ID},
			slog.String(c.HeadingIDKey, newHeading.ID),
		)
	}
}

func (h *HeadingHandler) GetHeadingByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.handlers.GetHeadingByID"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		headingID := chi.URLParam(r, c.HeadingIDKey)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryHeadingID)
			return
		}

		heading, err := h.headingStorage.GetHeadingByID(r.Context(), headingID, userID)
		switch {
		case errors.Is(err, c.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading received", heading, slog.String(c.HeadingIDKey, headingID))
		}
	}
}

func (h *HeadingHandler) GetHeadingsByListID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.handlers.GetHeadingsByListID"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
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

		headings, err := h.headingStorage.GetHeadingsByListID(r.Context(), listID, userID)
		switch {
		case errors.Is(err, c.ErrNoHeadingsFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoHeadingsFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetHeadingsByListID, err)
			return
		default:
			handleResponseSuccess(w, r, log, "headings found", headings, slog.Int(c.CountKey, len(headings)))
		}
	}
}

func (h *HeadingHandler) UpdateHeading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.handlers.UpdateHeading"

		log := logger.LogWithRequest(h.logger, op, r)
		heading := &models.Heading{}

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		headingID := chi.URLParam(r, c.HeadingIDKey)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryHeadingID)
			return
		}

		if err = DecodeAndValidateJSON(w, r, log, heading); err != nil {
			return
		}

		updatedHeading := models.Heading{
			ID:     headingID,
			Title:  heading.Title,
			UserID: userID,
		}

		err = h.headingStorage.UpdateHeading(r.Context(), updatedHeading)
		switch {
		case errors.Is(err, c.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToUpdateHeading, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading updated", updatedHeading, slog.String(c.HeadingIDKey, headingID))
		}
	}
}

func (h *HeadingHandler) MoveHeadingToAnotherList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.handlers.MoveTaskToAnotherList"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		headingID := chi.URLParam(r, c.HeadingIDKey)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryHeadingID)
			return
		}

		otherListID := r.URL.Query().Get(c.ListIDKey)
		if otherListID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryListID)
			return
		}

		err = h.headingStorage.MoveHeadingToAnotherList(r.Context(), headingID, otherListID, userID)
		switch {
		case errors.Is(err, c.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToMoveHeading, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading moved to another list", nil, slog.String(c.HeadingIDKey, headingID))
		}
	}
}

func (h *HeadingHandler) DeleteHeading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "heading.handlers.DeleteHeading"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		headingID := chi.URLParam(r, c.HeadingIDKey)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryHeadingID)
			return
		}

		err = h.headingStorage.DeleteHeading(r.Context(), headingID, userID)
		switch {
		case errors.Is(err, c.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToDeleteHeading, err)
			return
		default:
			handleResponseSuccess(w, r, log, "heading deleted", nil, slog.String(c.HeadingIDKey, headingID))
		}
	}
}
