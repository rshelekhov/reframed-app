package v1

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/rshelekhov/reframed/internal/config"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/port"
)

type AppRouter struct {
	*config.ServerSettings
	*slog.Logger
	*jwtoken.TokenService
	*authHandler
	*listHandler
	*headingHandler
	*taskHandler
	*tagHandler
	*statusHandler
}

func NewRouter(
	cfg *config.ServerSettings,
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	authUsecase port.AuthUsecase,
	listUsecase port.ListUsecase,
	headingUsecase port.HeadingUsecase,
	taskUsecase port.TaskUsecase,
	tagUsecase port.TagUsecase,
	statusUsecase port.StatusUsecase,
) *chi.Mux {
	ar := &AppRouter{
		ServerSettings: cfg,
		Logger:         log,
		TokenService:   jwt,
		authHandler:    newAuthHandler(log, jwt, authUsecase),
		listHandler:    newListHandler(log, jwt, listUsecase),
		headingHandler: newHeadingHandler(log, jwt, headingUsecase),
		taskHandler:    newTaskHandler(log, jwt, taskUsecase),
		tagHandler:     newTagHandler(log, jwt, tagUsecase),
		statusHandler:  newStatusHandler(log, jwt, statusUsecase),
	}

	return ar.initRoutes()
}
