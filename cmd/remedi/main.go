package main

import (
	"github.com/rshelekhov/remedi/internal/config"
	"github.com/rshelekhov/remedi/internal/lib/logger/sl"
	"log/slog"
)

func main() {
	cfg := config.MustLoad()

	log := sl.SetupLogger(cfg.Env)

	// A field with information about the current environment will be added to each message
	log = log.With(slog.String("env", cfg.Env))

	log.Info("initializing server", slog.String("address", cfg.Address))
	log.Debug("logger debug mode enabled")
}
