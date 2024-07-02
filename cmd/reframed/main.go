// Package main configures and runs application.
package main

import (
	"context"
	"github.com/rshelekhov/reframed/internal/config"
	"log/slog"

	"github.com/rshelekhov/reframed/internal/app/httpserver"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	ssogrpc "github.com/rshelekhov/reframed/internal/clients/sso/grpc"
	v1 "github.com/rshelekhov/reframed/internal/handler/http/v1"
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
		log,
		cfg.Clients.SSO.Address,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)

	// TODO: research where and how to set appID
	var appID int32
	appID = 1
	tokenAuth := jwtoken.NewService(ssoClient, appID)

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
	statusStorage := postgres.NewStatusStorage(pg)

	// Usecases
	authUsecase := usecase.NewAuthUsecase(ssoClient, tokenAuth)
	headingUsecase := usecase.NewHeadingUsecase(headingStorage)
	listUsecase := usecase.NewListUsecase(listStorage)
	tagUsecase := usecase.NewTagUsecase(tagStorage)
	taskUsecase := usecase.NewTaskUsecase(taskStorage)
	statusUsecase := usecase.NewStatusUsecase(statusStorage)

	authUsecase.ListUsecase = listUsecase
	authUsecase.HeadingUsecase = headingUsecase
	headingUsecase.ListUsecase = listUsecase
	listUsecase.HeadingUsecase = headingUsecase
	taskUsecase.HeadingUsecase = headingUsecase
	taskUsecase.TagUsecase = tagUsecase
	taskUsecase.ListUsecase = listUsecase

	// HTTP Server
	log.Info("starting httpserver", slog.String("address", cfg.HTTPServer.Address))

	router := v1.NewRouter(
		cfg,
		log,
		tokenAuth,
		authUsecase,
		listUsecase,
		headingUsecase,
		taskUsecase,
		tagUsecase,
		statusUsecase,
	)

	srv := httpserver.NewServer(cfg, log, tokenAuth, router)
	srv.Start()
}
