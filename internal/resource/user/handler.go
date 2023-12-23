package user

import (
	"database/sql"
	"errors"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	resp "github.com/rshelekhov/remedi/internal/lib/api/response"
	"github.com/rshelekhov/remedi/internal/lib/logger/sl"
	"github.com/rshelekhov/remedi/internal/service"
	"io"
	"log/slog"
	"net/http"
)

type Handler struct {
	logger  *slog.Logger
	storage *Storage
	service Service
}

func New(log *slog.Logger, db *sql.DB) *Handler {
	return &Handler{
		logger:  log,
		storage: NewStorage(db),
	}
}

type HandlerUpd struct {
	logger  *slog.Logger
	service service.UserService
}

func NewHandlerUpd(
	log *slog.Logger,
	service service.UserService,
) *HandlerUpd {
	return &HandlerUpd{
		logger:  log,
		service: service,
	}
}

func (h *Handler) ListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handler.ListUsers"
	}
}

func (hu *HandlerUpd) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handler.CreateUser"

		log := hu.logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var user CreateUser

		err := render.DecodeJSON(r.Body, &user)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("request body is empty"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request body"))

			return
		}

		log.Info("request body decoded", slog.Any("user", user))

		/*id, err := hu.service.CreateUser(user)
		if err != nil {
			log.Error("failed to create user", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to create user"))

			return
		}

		log.Info("User created", slog.Any("user_id", id))

		render.JSON(w, r, resp.Success("User created", id))*/
	}
}

func (h *Handler) ReadUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handler.ReadUser"
	}
}

func (h *Handler) UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handler.UpdateUser"
	}
}

func (h *Handler) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "user.handler.DeleteUser"
	}
}
