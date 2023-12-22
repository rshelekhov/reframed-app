package handlers

import (
	"github.com/go-chi/chi"
	"github.com/rshelekhov/remedi/internal/service"
	"github.com/rshelekhov/remedi/internal/storage/postgres"
	"log/slog"
)

type Resource struct {
	logger  *slog.Logger
	router  *chi.Mux
	service service.Service
}

func NewResource(logger *slog.Logger, router *chi.Mux, storage postgres.Storage) *Resource {
	return &Resource{
		logger:  logger,
		router:  router,
		service: service.NewService(storage),
	}
}
