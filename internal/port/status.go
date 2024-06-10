package port

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
)

type (
	StatusUsecase interface {
		GetStatuses(ctx context.Context) ([]model.StatusResponseData, error)
		GetStatusID(ctx context.Context, statusName string) (int, error)
		GetStatusName(ctx context.Context, statusID int) (string, error)
	}

	StatusStorage interface {
	}
)
