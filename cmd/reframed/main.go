// Package main configures and runs application.
package main

import (
	"context"
	"github.com/rshelekhov/reframed/internal/config"
	"log/slog"

	"github.com/rshelekhov/reframed/internal/app/httpserver"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	ssogrpc "github.com/rshelekhov/reframed/internal/clients/sso/grpc"
	v1 "github.com/rshelekhov/reframed/internal/controller/http/v1"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/storage/postgres"
	"github.com/rshelekhov/reframed/internal/usecase"
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

	ssoClient, err := ssogrpc.New(
		context.Background(),
		log.Logger,
		cfg.Clients.SSO.Address,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)

	// TODO: research where and how to set appID
	var appID int32
	appID = 1
	tokenAuth := jwtoken.NewJWTokenService(ssoClient, appID)

	// Storage
	pg, err := postgres.NewStorage(cfg)
	if err != nil {
		log.Error("failed to init storage", logger.Err(err))
	}

	log.Debug("storage initiated")

	headingStorage := postgres.NewHeadingStorage(pg)
	listStorage := postgres.NewListStorage(pg)
	taskStorage := postgres.NewTaskStorage(pg)
	tagStorage := postgres.NewTagStorage(pg)

	// Usecases
	headingUsecase := usecase.NewHeadingUsecase(headingStorage)
	listUsecase := usecase.NewListUsecase(listStorage, headingUsecase)
	authUsecase := usecase.NewAuthUsecase(ssoClient, tokenAuth, listUsecase, headingUsecase)
	tagUsecase := usecase.NewTagUsecase(tagStorage)
	taskUsecase := usecase.NewTaskUsecase(taskStorage, headingUsecase, tagUsecase, listUsecase)

	// HTTP Server
	log.Info("starting httpserver", slog.String("address", cfg.HTTPServer.Address))

	router := v1.NewRouter(
		log.Logger,
		tokenAuth,
		authUsecase,
		listUsecase,
		headingUsecase,
		taskUsecase,
		tagUsecase,
	)

	srv := httpserver.NewServer(cfg, log.Logger, tokenAuth, router)
	srv.Start()
}
