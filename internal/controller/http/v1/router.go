package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/port"
	"log/slog"
)

type AppRouter struct {
	*slog.Logger
	*jwtoken.TokenService
	*authController
	*listController
	*headingController
	*taskController
	*tagController
}

func NewRouter(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	authUsecase port.AuthUsecase,
	listUsecase port.ListUsecase,
	headingUsecase port.HeadingUsecase,
	taskUsecase port.TaskUsecase,
	tagUsecase port.TagUsecase,
) *chi.Mux {
	ar := &AppRouter{
		Logger:            log,
		TokenService:      jwt,
		authController:    newAuthController(log, jwt, authUsecase),
		listController:    newListController(log, jwt, listUsecase),
		headingController: newHeadingController(log, jwt, headingUsecase),
		taskController:    newTaskController(log, jwt, taskUsecase),
		tagController:     newTagController(log, jwt, tagUsecase),
	}

	return ar.initRoutes()
}
