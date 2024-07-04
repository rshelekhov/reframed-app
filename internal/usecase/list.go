package usecase

import (
	"context"
	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type ListUsecase struct {
	storage        port.ListStorage
	HeadingUsecase port.HeadingUsecase
}

func NewListUsecase(listStorage port.ListStorage) *ListUsecase {
	return &ListUsecase{storage: listStorage}
}

func (u *ListUsecase) CreateList(ctx context.Context, data *model.ListRequestData) (model.ListResponseData, error) {
	currentTime := time.Now()

	newList := model.List{
		ID:        ksuid.New().String(),
		Title:     data.Title,
		IsDefault: false,
		UserID:    data.UserID,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	if err := u.storage.CreateList(ctx, newList); err != nil {
		return model.ListResponseData{}, err
	}

	defaultHeading := model.Heading{
		ID:        ksuid.New().String(),
		Title:     model.DefaultHeading.String(),
		ListID:    newList.ID,
		UserID:    data.UserID,
		IsDefault: true,
		UpdatedAt: currentTime,
	}

	if err := u.HeadingUsecase.CreateDefaultHeading(ctx, defaultHeading); err != nil {
		return model.ListResponseData{}, err
	}

	return model.ListResponseData{
		ID:        newList.ID,
		Title:     newList.Title,
		UserID:    newList.UserID,
		CreatedAt: newList.CreatedAt,
		UpdatedAt: newList.UpdatedAt,
	}, nil
}

func (u *ListUsecase) CreateDefaultList(ctx context.Context, userID string) error {
	currentTime := time.Now()

	defaultList := model.List{
		ID:        ksuid.New().String(),
		Title:     model.DefaultInboxList.String(),
		IsDefault: true,
		UserID:    userID,
		UpdatedAt: currentTime,
	}

	if err := u.storage.CreateList(ctx, defaultList); err != nil {
		return err
	}

	defaultHeading := model.Heading{
		ID:        ksuid.New().String(),
		Title:     model.DefaultHeading.String(),
		ListID:    defaultList.ID,
		UserID:    userID,
		IsDefault: true,
		UpdatedAt: currentTime,
	}

	if err := u.HeadingUsecase.CreateDefaultHeading(ctx, defaultHeading); err != nil {
		return err
	}

	return nil
}

func (u *ListUsecase) GetListByID(ctx context.Context, data model.ListRequestData) (model.ListResponseData, error) {
	list, err := u.storage.GetListByID(ctx, data.ID, data.UserID)
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
	lists, err := u.storage.GetListsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var listResp []model.ListResponseData

	for _, list := range lists {
		listResp = append(listResp, mapListToResponseData(list))
	}

	return listResp, nil
}

func (u *ListUsecase) GetDefaultListID(ctx context.Context, userID string) (string, error) {
	listID, err := u.storage.GetDefaultListID(ctx, userID)
	if err != nil {
		return "", err
	}

	return listID, nil
}

func mapListToResponseData(list model.List) model.ListResponseData {
	return model.ListResponseData{
		ID:        list.ID,
		Title:     list.Title,
		UserID:    list.UserID,
		UpdatedAt: list.UpdatedAt,
	}
}

func (u *ListUsecase) UpdateList(ctx context.Context, data *model.ListRequestData) (model.ListResponseData, error) {
	updatedList := model.List{
		ID:        data.ID,
		Title:     data.Title,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	if err := u.storage.UpdateList(ctx, updatedList); err != nil {
		return model.ListResponseData{}, err
	}

	return model.ListResponseData{
		ID:        updatedList.ID,
		Title:     updatedList.Title,
		UserID:    updatedList.UserID,
		UpdatedAt: updatedList.UpdatedAt,
	}, nil
}

func (u *ListUsecase) DeleteList(ctx context.Context, data model.ListRequestData) error {
	// Check if list is not default list
	list, err := u.storage.GetListByID(ctx, data.ID, data.UserID)
	if err != nil {
		return err
	}
	if list.IsDefault {
		return le.ErrCannotDeleteDefaultList
	}

	deletedList := model.List{
		ID:        data.ID,
		UserID:    data.UserID,
		DeletedAt: time.Now(),
	}

	return u.storage.DeleteList(ctx, deletedList)
}
