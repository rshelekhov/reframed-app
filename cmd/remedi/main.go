package main

import (
	"github.com/rshelekhov/remedi/api/router"
	"github.com/rshelekhov/remedi/internal/config"
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
	
	r := router.New(log)

	// TODO: remove this
	defer func(storage *postgres.Storage) {
		err := storage.Close()
		if err != nil {
			log.Error("failed to close storage", err)
			os.Exit(1)
		}
	}(storage)
}
