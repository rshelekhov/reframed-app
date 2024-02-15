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

type listController struct {
	logger  logger.Interface
	jwt     *jwtoken.TokenService
	usecase domain.ListUsecase
}

func NewListRoutes(
	r *chi.Mux,
	log logger.Interface,
	jwt *jwtoken.TokenService,
	usecase domain.ListUsecase,
) {
	c := &listController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtoken.Verifier(jwt))
		r.Use(jwtoken.Authenticator())

		r.Route("/user/lists", func(r chi.Router) {
			r.Get("/", c.GetListsByUserID())
			r.Post("/", c.CreateList())

			r.Route("/{list_id}", func(r chi.Router) {
				r.Get("/", c.GetListByID())
				r.Put("/", c.UpdateList())
				r.Delete("/", c.DeleteList())
			})
		})
	})
}

func (c *listController) CreateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.controller.CreateList"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		listInput := &domain.ListRequestData{}
		if err := decodeAndValidateJSON(w, r, log, listInput); err != nil {
			return
		}

		listInput.UserID = userID

		listID, err := c.usecase.CreateList(ctx, listInput)
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToCreateList, err)
			return
		}

		handleResponseCreated(
			w, r, log, "list created",
			domain.ListResponseData{ID: listID},
			slog.String(key.ListID, listID),
		)
	}
}

func (c *listController) GetListByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.controller.GetListByID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryListID)
			return
		}

		listInput := domain.ListRequestData{
			ID:     listID,
			UserID: userID,
		}

		listResp, err := c.usecase.GetListByID(ctx, listInput)
		switch {
		case errors.Is(err, domain.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		listsResp, err := c.usecase.GetListsByUserID(ctx, userID)
		switch {
		case errors.Is(err, domain.ErrNoListsFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoListsFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetLists, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryListID)
			return
		}

		listInput := &domain.ListRequestData{}
		if err := decodeAndValidateJSON(w, r, log, listInput); err != nil {
			return
		}

		listInput.ID = listID
		listInput.UserID = userID

		err := c.usecase.UpdateList(ctx, listInput)
		switch {
		case errors.Is(err, domain.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToUpdateList, err)
			return
		default:
			handleResponseSuccess(w, r, log, "list updated", listID, slog.String(key.ListID, listID))
		}
	}
}

func (c *listController) DeleteList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "list.controller.DeleteList"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryListID)
			return
		}

		listInput := domain.ListRequestData{
			ID:     listID,
			UserID: userID,
		}

		err := c.usecase.DeleteList(ctx, listInput)
		switch {
		case errors.Is(err, domain.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrListNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToDeleteList, err)
			return
		default:
			handleResponseSuccess(w, r, log, "list deleted", domain.ListResponseData{ID: listID}, slog.String(key.ListID, listID))
		}
	}
}
