package router

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/jmoiron/sqlx"
	mwlogger "github.com/rshelekhov/remedi/internal/http-server/middleware/logger"
	"github.com/rshelekhov/remedi/internal/resource/health"
	userHandlers "github.com/rshelekhov/remedi/internal/resource/user"
	"log/slog"
	"time"
)

func New(log *slog.Logger, db *sqlx.DB, validate *validator.Validate) *chi.Mux {
	r := chi.NewRouter()

	// Add request_id to each request, for tracing purposes
	r.Use(middleware.RequestID)

	// Logging of all requests
	r.Use(middleware.Logger)

	// By default, middleware.Logger uses its own internal logger,
	// which should be overridden to use ours. Otherwise, problems
	// may arise - for example, with log collection. We can use
	// our own middleware to log requests:
	r.Use(mwlogger.New(log))

	// If a panic happens somewhere inside the server (request handlers),
	// the application should not crash.
	r.Use(middleware.Recoverer)

	// Parser of incoming request URLs
	r.Use(middleware.URLFormat)

	// Set the content type to application/json
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Enable httprate request limiter of 100 requests per minute per IP
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	// Health check
	r.Get("/health", health.Read())

	// Handlers
	userHandlers.Activate(r, log, db, validate)

	return r
}
