package main

import (
	"fmt"
	"github.com/rshelekhov/remedi/internal/config"
	"github.com/rshelekhov/remedi/internal/lib/logger/sl"
	"github.com/rshelekhov/remedi/internal/storage/postgres"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustLoad()

	log := sl.SetupLogger(cfg.Env)

	// A field with information about the current environment will be added to each message
	log = log.With(slog.String("env", cfg.Env))

	log.Info(
		"initializing server",
		slog.String("address", cfg.HTTPServer.Address))
	log.Debug("logger debug mode enabled")

	storage, err := postgres.NewStorage(
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Postgres.Host,
			cfg.Postgres.Port,
			cfg.Postgres.User,
			cfg.Postgres.Password,
			cfg.Postgres.DBName,
			cfg.Postgres.SSLMode))
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
	}

	defer func(storage *postgres.Storage) {
		err := storage.Close()
		if err != nil {
			log.Error("failed to close storage", err)
			os.Exit(1)
		}
	}(storage)
}
