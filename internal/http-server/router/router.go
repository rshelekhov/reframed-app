package router

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rshelekhov/remedi/internal/http-server/handlers"
	"github.com/rshelekhov/remedi/internal/http-server/handlers/health"
	"github.com/rshelekhov/remedi/internal/storage/postgres"
	"log/slog"
)

func New(log *slog.Logger, storage postgres.Storage) *chi.Mux {
	r := chi.NewRouter()

	// Add request_id to each request, for tracing purposes
	r.Use(middleware.RequestID)

	// Logging of all requests
	r.Use(middleware.Logger)

	// If a panic happens somewhere inside the server (request handler),
	// the application should not crash.
	r.Use(middleware.Recoverer)

	// Parser of incoming request URLs
	r.Use(middleware.URLFormat)

	handler := handlers.NewHandler(log, r, storage)

	// By default, middleware.Logger uses its own internal logger,
	// which should be overridden to use ours. Otherwise, problems
	// may arise - for example, with log collection. We can use
	// our own middleware to log requests:
	r.Use(handler.MiddlewareLogger(log))

	r.Use(render.SetContentType(render.ContentTypeJSON))

	health.RegisterHandlers(r)

	return r
}
