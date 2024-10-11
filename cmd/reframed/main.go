// Package main configures and runs application.
package main

import (
	"log/slog"

	"github.com/rshelekhov/jwtauth"
	"github.com/rshelekhov/reframed/internal/config"

	"github.com/rshelekhov/reframed/internal/app/httpserver"

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
		log,
		cfg.Clients.SSO.Address,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)
	if err != nil {
		log.Error("failed to init sso client", logger.Err(err))
	}

	// tokenAuth := jwtoken.NewService(ssoClient, cfg.AppData.ID)
	tokenAuth := jwtauth.New(ssoClient, cfg.AppData.ID)

	// Storage
	pg, err := postgres.NewStorage(cfg)
	if err != nil {
		log.Error("failed to init storage", logger.Err(err))
	}

	log.Debug("storage initiated")

	userStorage := postgres.NewUserStorage(pg)
	headingStorage := postgres.NewHeadingStorage(pg)
	listStorage := postgres.NewListStorage(pg)
	taskStorage := postgres.NewTaskStorage(pg)
	tagStorage := postgres.NewTagStorage(pg)
	statusStorage := postgres.NewStatusStorage(pg)

	// Usecases
	userUsecase := usecase.NewUserUsecase(userStorage)
	authUsecase := usecase.NewAuthUsecase(cfg, ssoClient, tokenAuth)
	headingUsecase := usecase.NewHeadingUsecase(headingStorage)
	listUsecase := usecase.NewListUsecase(listStorage)
	tagUsecase := usecase.NewTagUsecase(tagStorage)
	taskUsecase := usecase.NewTaskUsecase(taskStorage)
	statusUsecase := usecase.NewStatusUsecase(statusStorage)

	authUsecase.UserUsecase = userUsecase
	authUsecase.ListUsecase = listUsecase
	authUsecase.HeadingUsecase = headingUsecase
	headingUsecase.ListUsecase = listUsecase
	headingUsecase.TaskUsecase = taskUsecase
	listUsecase.HeadingUsecase = headingUsecase
	listUsecase.TaskUsecase = taskUsecase
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
