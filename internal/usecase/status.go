package usecase

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type StatusUsecase struct {
	storage port.StatusStorage
}

func NewStatusUsecase(storage port.StatusStorage) *StatusUsecase {
	return &StatusUsecase{storage: storage}
}

func (u *StatusUsecase) GetStatuses(ctx context.Context) ([]model.StatusResponseData, error) {
	statuses, err := u.storage.GetStatuses(ctx)
	if err != nil {
		return nil, err
	}

	var statusesResp []model.StatusResponseData

	for _, status := range statuses {
		statusesResp = append(statusesResp, model.StatusResponseData{
			ID:    int(status.ID),
			Title: status.Title,
		})
	}

	return statusesResp, nil
}

func (u *StatusUsecase) GetStatusByID(ctx context.Context, statusID int) (model.StatusResponseData, error) {
	statusName, err := u.storage.GetStatusByID(ctx, int32(statusID))
	if err != nil {
		return model.StatusResponseData{}, err
	}

	return model.StatusResponseData{
		ID:    statusID,
		Title: statusName,
	}, nil
}
