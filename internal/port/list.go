package port

import (
	"context"

	"github.com/rshelekhov/reframed/internal/model"
)

type (
	ListUsecase interface {
		CreateList(ctx context.Context, data *model.ListRequestData) (model.ListResponseData, error)
		CreateDefaultList(ctx context.Context, userID string) error
		GetListByID(ctx context.Context, data model.ListRequestData) (model.ListResponseData, error)
		GetListsByUserID(ctx context.Context, userID string) ([]model.ListResponseData, error)
		GetDefaultListID(ctx context.Context, userID string) (string, error)
		UpdateList(ctx context.Context, data *model.ListRequestData) (model.ListResponseData, error)
		DeleteList(ctx context.Context, data model.ListRequestData) error
	}

	ListStorage interface {
		CreateList(ctx context.Context, list model.List) error
		GetListByID(ctx context.Context, listID, userID string) (model.List, error)
		GetListsByUserID(ctx context.Context, userID string) ([]model.List, error)
		GetDefaultListID(ctx context.Context, userID string) (string, error)
		UpdateList(ctx context.Context, list model.List) error
		DeleteList(ctx context.Context, list model.List) error
	}
)
