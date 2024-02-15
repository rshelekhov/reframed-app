package usecase

import (
	"context"
	"github.com/rshelekhov/reframed/internal/domain"
	"github.com/segmentio/ksuid"
	"time"
)

type ListUsecase struct {
	listStorage    domain.ListStorage
	headingUsecase domain.HeadingUsecase
}

func NewListUsecase(listStorage domain.ListStorage, headingUsecase domain.HeadingUsecase) *ListUsecase {
	return &ListUsecase{
		listStorage:    listStorage,
		headingUsecase: headingUsecase,
	}
}

func (u *ListUsecase) CreateList(ctx context.Context, data *domain.ListRequestData) (string, error) {
	updatedAt := time.Now().UTC()

	newList := domain.List{
		ID:        ksuid.New().String(),
		Title:     data.Title,
		UserID:    data.UserID,
		UpdatedAt: &updatedAt,
	}

	if err := u.listStorage.CreateList(ctx, newList); err != nil {
		return "", err
	}

	defaultHeading := domain.Heading{
		ID:        ksuid.New().String(),
		Title:     domain.DefaultHeading.String(),
		ListID:    newList.ID,
		UserID:    data.UserID,
		IsDefault: true,
		UpdatedAt: &updatedAt,
	}

	if err := u.headingUsecase.CreateDefaultHeading(ctx, defaultHeading); err != nil {
		return "", err
	}

	return newList.ID, nil
}

func (u *ListUsecase) CreateDefaultList(ctx context.Context, list domain.List) error {
	return u.listStorage.CreateList(ctx, list)
}

func (u *ListUsecase) GetListByID(ctx context.Context, data domain.ListRequestData) (domain.ListResponseData, error) {
	list, err := u.listStorage.GetListByID(ctx, data.ID, data.UserID)
	if err != nil {
		return domain.ListResponseData{}, err
	}

	return domain.ListResponseData{
		ID:        list.ID,
		Title:     list.Title,
		UserID:    list.UserID,
		UpdatedAt: list.UpdatedAt,
	}, nil
}

func (u *ListUsecase) GetListsByUserID(ctx context.Context, userID string) ([]domain.ListResponseData, error) {
	lists, err := u.listStorage.GetListsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var listResp []domain.ListResponseData

	for _, list := range lists {
		listResp = append(listResp, mapListToResponseData(list))
	}

	return listResp, nil
}

func mapListToResponseData(list domain.List) domain.ListResponseData {
	return domain.ListResponseData{
		ID:        list.ID,
		Title:     list.Title,
		UserID:    list.UserID,
		UpdatedAt: list.UpdatedAt,
	}
}

func (u *ListUsecase) UpdateList(ctx context.Context, data *domain.ListRequestData) error {
	updatedAt := time.Now().UTC()

	updatedList := domain.List{
		ID:        data.ID,
		Title:     data.Title,
		UserID:    data.UserID,
		UpdatedAt: &updatedAt,
	}

	return u.listStorage.UpdateList(ctx, updatedList)
}

func (u *ListUsecase) DeleteList(ctx context.Context, data domain.ListRequestData) error {
	deletedAt := time.Now().UTC()

	deletedList := domain.List{
		ID:        data.ID,
		UserID:    data.UserID,
		DeletedAt: &deletedAt,
	}

	return u.listStorage.DeleteList(ctx, deletedList)
}
