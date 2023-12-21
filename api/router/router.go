package router

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	mwlogger "github.com/rshelekhov/remedi/internal/http-server/middleware/logger"
	"log/slog"
)

func New(log *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	// Add request_id to each request, for tracing purposes
	router.Use(middleware.RequestID)

	// Logging of all requests
	router.Use(middleware.Logger)

	// By default, middleware.Logger uses its own internal logger,
	// which should be overridden to use ours. Otherwise, problems
	// may arise - for example, with log collection. We can use
	// our own middleware to log requests:
	router.Use(mwlogger.New(log))

	// If a panic happens somewhere inside the server (request handler),
	// the application should not crash.
	router.Use(middleware.Recoverer)

	// Parser of incoming request URLs
	router.Use(middleware.URLFormat)

	router.Use(render.SetContentType(render.ContentTypeJSON))

	return router
}
