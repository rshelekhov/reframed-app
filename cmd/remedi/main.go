package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rshelekhov/remedi/internal/config"
	mwlogger "github.com/rshelekhov/remedi/internal/http-server/middleware/logger"
	"github.com/rshelekhov/remedi/internal/storage/postgres"
	"github.com/rshelekhov/remedi/internal/util/logger/sl"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustLoad()

	log := sl.SetupLogger(cfg.AppEnv)

	// A field with information about the current environment
	// will be added to each message
	log = log.With(slog.String("env", cfg.AppEnv))

	log.Info(
		"initializing server",
		slog.String("address", cfg.HTTPServer.Address))
	log.Debug("logger debug mode enabled")

	storage, err := postgres.NewStorage(cfg.Postgres.URL)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
	}
	log.Debug("storage initiated")

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

	// TODO: remove this
	defer func(storage *postgres.Storage) {
		err := storage.Close()
		if err != nil {
			log.Error("failed to close storage", err)
			os.Exit(1)
		}
	}(storage)
}
