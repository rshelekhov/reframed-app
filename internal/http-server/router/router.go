package router

import (
	"database/sql"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	mwlogger "github.com/rshelekhov/remedi/internal/http-server/middleware/logger"
	"github.com/rshelekhov/remedi/internal/resource/health"
	"github.com/rshelekhov/remedi/internal/resource/user"
	"log/slog"
)

func New(log *slog.Logger, db *sql.DB) *chi.Mux {
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

	// If a panic happens somewhere inside the server (request handler),
	// the application should not crash.
	r.Use(middleware.Recoverer)

	// Parser of incoming request URLs
	r.Use(middleware.URLFormat)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Health check
	r.Get("/health", health.Read())

	userHandler := user.New(log, db)
	r.Get("/users", userHandler.ListUsers())
	r.Get("/users/{id}", userHandler.ReadUser())
	r.Put("/users/{id}", userHandler.UpdateUser())
	r.Delete("/users/{id}", userHandler.DeleteUser())

	userStorage := user.NewStorage(db)
	userService := user.NewServiceUpd(userStorage)
	user := user.NewHandlerUpd(log, userService)

	r.Post("/users", user.CreateUser())

	return r
}
