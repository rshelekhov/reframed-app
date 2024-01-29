// Package main configures and runs application.
package main

import (
	"github.com/rshelekhov/reframed/config"
	"github.com/rshelekhov/reframed/internal/handlers"
	"github.com/rshelekhov/reframed/internal/http-server"
	"github.com/rshelekhov/reframed/internal/http-server/middleware/auth"
	"github.com/rshelekhov/reframed/internal/logger"
	"github.com/rshelekhov/reframed/internal/storage/postgres"
	"log/slog"
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

	tokenAuth := auth.NewJWTAuth(
		cfg.JWTAuth.Secret,
		cfg.JWTAuth.AccessTokenTTL,
		cfg.JWTAuth.RefreshTokenTTL,
		cfg.JWTAuth.RefreshTokenCookiePath,
	)

	// Storage
	pg, err := postgres.NewStorage(cfg)
	if err != nil {
		log.Error("failed to init storage", logger.Err(err))
	}
	log.Debug("storage initiated")

	userStorage := postgres.NewUserStorage(pg)
	listStorage := postgres.NewListStorage(pg)

	// Handlers
	user := handlers.NewUserHandler(log, tokenAuth, userStorage, listStorage)
	list := handlers.NewListHandler(log, tokenAuth, listStorage)

	// Routers
	/*route.NewUserRouter(r, log, tokenAuth, userStorage, listStorage)
	route.NewListRouter(r, log, tokenAuth, listStorage)*/

	// HTTP Server
	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := http_server.NewServer(cfg, log, tokenAuth, user, list)
	srv.Start()
}
