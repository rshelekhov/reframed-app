package usecase

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
	"github.com/segmentio/ksuid"
	"time"
)

type ListUsecase struct {
	listStorage    port.ListStorage
	headingUsecase port.HeadingUsecase
}

func NewListUsecase(listStorage port.ListStorage, headingUsecase port.HeadingUsecase) *ListUsecase {
	return &ListUsecase{
		listStorage:    listStorage,
		headingUsecase: headingUsecase,
	}
}

func (u *ListUsecase) CreateList(ctx context.Context, data *model.ListRequestData) (string, error) {
	newList := model.List{
		ID:        ksuid.New().String(),
		Title:     data.Title,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	if err := u.listStorage.CreateList(ctx, newList); err != nil {
		return "", err
	}

	defaultHeading := model.Heading{
		ID:        ksuid.New().String(),
		Title:     model.DefaultHeading.String(),
		ListID:    newList.ID,
		UserID:    data.UserID,
		IsDefault: true,
		UpdatedAt: time.Now(),
	}

	if err := u.headingUsecase.CreateDefaultHeading(ctx, defaultHeading); err != nil {
		return "", err
	}

	return newList.ID, nil
}

func (u *ListUsecase) CreateDefaultList(ctx context.Context, list model.List) error {
	return u.listStorage.CreateList(ctx, list)
}

func (u *ListUsecase) GetListByID(ctx context.Context, data model.ListRequestData) (model.ListResponseData, error) {
	list, err := u.listStorage.GetListByID(ctx, data.ID, data.UserID)
	if err != nil {
		return model.ListResponseData{}, err
	}

	return model.ListResponseData{
		ID:        list.ID,
		Title:     list.Title,
		UserID:    list.UserID,
		UpdatedAt: list.UpdatedAt,
	}, nil
}

func (u *ListUsecase) GetListsByUserID(ctx context.Context, userID string) ([]model.ListResponseData, error) {
	lists, err := u.listStorage.GetListsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var listResp []model.ListResponseData

	for _, list := range lists {
		listResp = append(listResp, mapListToResponseData(list))
	}

	return listResp, nil
}

func mapListToResponseData(list model.List) model.ListResponseData {
	return model.ListResponseData{
		ID:        list.ID,
		Title:     list.Title,
		UserID:    list.UserID,
		UpdatedAt: list.UpdatedAt,
	}
}

func (u *ListUsecase) UpdateList(ctx context.Context, data *model.ListRequestData) error {
	updatedList := model.List{
		ID:        data.ID,
		Title:     data.Title,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	return u.listStorage.UpdateList(ctx, updatedList)
}

func (u *ListUsecase) DeleteList(ctx context.Context, data model.ListRequestData) error {
	deletedList := model.List{
		ID:        data.ID,
		UserID:    data.UserID,
		DeletedAt: time.Now(),
	}

	return u.listStorage.DeleteList(ctx, deletedList)
}
