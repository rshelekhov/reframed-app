package port

import (
	"context"

	"github.com/rshelekhov/reframed/internal/model"
)

type (
	StatusUsecase interface {
		GetStatuses(ctx context.Context) ([]model.StatusResponseData, error)
		GetStatusByID(ctx context.Context, statusID int) (model.StatusResponseData, error)
	}

	StatusStorage interface {
		GetStatuses(ctx context.Context) ([]model.Status, error)
		GetStatusByID(ctx context.Context, statusID int32) (string, error)
	}
)
