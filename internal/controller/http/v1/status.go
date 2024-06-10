package v1

import (
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/port"
	"log/slog"
	"net/http"
)

type statusController struct {
	logger  *slog.Logger
	jwt     *jwtoken.TokenService
	usecase port.StatusUsecase
}

func newStatusController(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.StatusUsecase,
) *statusController {
	return &statusController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

func (c *statusController) GetStatuses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (c *statusController) GetStatusID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (c *statusController) GetStatusName() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
