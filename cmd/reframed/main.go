// Package main configures and runs application.
package main

import (
	"github.com/rshelekhov/reframed/config"
	"github.com/rshelekhov/reframed/internal/api/route"
	"github.com/rshelekhov/reframed/internal/usecase"
	"github.com/rshelekhov/reframed/internal/usecase/storage"
	"github.com/rshelekhov/reframed/pkg/http-server"
	"github.com/rshelekhov/reframed/pkg/logger"
	"github.com/rshelekhov/reframed/pkg/storage/postgres"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.AppEnv)

	// A field with information about the current environment
	// will be added to each message
	log = log.With(slog.String("env", cfg.AppEnv))

	log.Info(
		"initializing server",
		slog.String("address", cfg.HTTPServer.Address))
	log.Debug("logger debug mode enabled")

	// Storage
	pg, err := postgres.NewPostgresStorage(cfg.Postgres.URL)
	if err != nil {
		log.Error("failed to init storage", logger.Err(err))
	}
	log.Debug("storage initiated")

	// Use cases
	userUsecase := usecase.NewUserUsecase(storage.NewUserStorage(pg))

	// Router
	r := route.NewRouter(log)

	// Routers
	route.NewUserRouter(r, log, userUsecase)

	// HTTP Server
	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := http_server.NewServer(cfg, log, r)
	srv.Start()

	defer func(pg *postgres.Storage) {
		err := pg.Close()
		if err != nil {
			log.Error("failed to close storage", err)
			os.Exit(1)
		}
	}(pg)
}
