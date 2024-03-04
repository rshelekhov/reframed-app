// Package main configures and runs application.
package main

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/rshelekhov/reframed/config"
	"github.com/rshelekhov/reframed/internal/controller/http/v1"
	"github.com/rshelekhov/reframed/internal/storage/postgres"
	"github.com/rshelekhov/reframed/internal/usecase"
	"github.com/rshelekhov/reframed/pkg/httpserver"
	"github.com/rshelekhov/reframed/pkg/httpserver/middleware/jwtoken"
	"github.com/rshelekhov/reframed/pkg/logger"
	"log/slog"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.AppEnv)

	// A field with information about the current environment
	// will be added to each message
	log = log.With(slog.String("env", cfg.AppEnv))

	log.Info(
		"initializing httpserver",
		slog.String("address", cfg.HTTPServer.Address))
	log.Debug("logger debug mode enabled")

	tokenAuth := jwtoken.NewJWTokenService(
		cfg.JWTAuth.SigningKey,
		jwt.SigningMethodHS256,
		cfg.JWTAuth.AccessTokenTTL,
		cfg.JWTAuth.RefreshTokenTTL,
		cfg.JWTAuth.RefreshTokenCookieDomain,
		cfg.JWTAuth.RefreshTokenCookiePath,
		cfg.JWTAuth.PasswordHash.Cost,
		cfg.JWTAuth.PasswordHash.Salt,
	)

	// Storage
	pg, err := postgres.NewStorage(cfg)
	if err != nil {
		log.Error("failed to init storage", logger.Err(err))
	}
	log.Debug("storage initiated")

	headingStorage := postgres.NewHeadingStorage(pg)
	listStorage := postgres.NewListStorage(pg)
	authStorage := postgres.NewAuthStorage(pg)
	taskStorage := postgres.NewTaskStorage(pg)
	tagStorage := postgres.NewTagStorage(pg)

	// Usecases
	headingUsecase := usecase.NewHeadingUsecase(headingStorage)
	listUsecase := usecase.NewListUsecase(listStorage, headingUsecase)
	authUsecase := usecase.NewAuthUsecase(authStorage, listUsecase, headingUsecase)
	tagUsecase := usecase.NewTagUsecase(tagStorage)
	taskUsecase := usecase.NewTaskUsecase(taskStorage, headingUsecase, tagUsecase, listUsecase)

	// HTTP Server
	log.Info("starting httpserver", slog.String("address", cfg.HTTPServer.Address))

	router := v1.NewRouter(
		log,
		tokenAuth,
		authUsecase,
		listUsecase,
		headingUsecase,
		taskUsecase,
		tagUsecase,
	)

	srv := httpserver.NewServer(cfg, log, tokenAuth, router)
	srv.Start()
}
