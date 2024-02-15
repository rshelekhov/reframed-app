package usecase

import (
	"context"
	"github.com/rshelekhov/reframed/internal/domain"
	"github.com/segmentio/ksuid"
	"time"
)

type HeadingUsecase struct {
	headingStorage domain.HeadingStorage
}

func NewHeadingUsecase(storage domain.HeadingStorage) *HeadingUsecase {
	return &HeadingUsecase{
		headingStorage: storage,
	}
}

func (u *HeadingUsecase) CreateHeading(ctx context.Context, data *domain.HeadingRequestData) (string, error) {
	updatedAt := time.Now().UTC()

	newHeading := domain.Heading{
		ID:        ksuid.New().String(),
		Title:     data.Title,
		ListID:    data.ListID,
		UserID:    data.UserID,
		IsDefault: false,
		UpdatedAt: &updatedAt,
	}

	if err := u.headingStorage.CreateHeading(ctx, newHeading); err != nil {
		return "", err
	}

	return newHeading.ID, nil
}

func (u *HeadingUsecase) CreateDefaultHeading(ctx context.Context, heading domain.Heading) error {
	return u.headingStorage.CreateHeading(ctx, heading)
}

func (u *HeadingUsecase) GetHeadingByID(ctx context.Context, data domain.HeadingRequestData) (domain.HeadingResponseData, error) {
	heading, err := u.headingStorage.GetHeadingByID(ctx, data.ID, data.UserID)
	if err != nil {
		return domain.HeadingResponseData{}, err
	}

	return domain.HeadingResponseData{
		ID:        heading.ID,
		Title:     heading.Title,
		ListID:    heading.ListID,
		UserID:    heading.UserID,
		UpdatedAt: heading.UpdatedAt,
	}, nil
}

func (u *HeadingUsecase) GetDefaultHeadingID(ctx context.Context, data domain.HeadingRequestData) (string, error) {
	return u.headingStorage.GetDefaultHeadingID(ctx, data.ListID, data.UserID)
}

func (u *HeadingUsecase) GetHeadingsByListID(ctx context.Context, data domain.HeadingRequestData) ([]domain.HeadingResponseData, error) {
	headings, err := u.headingStorage.GetHeadingsByListID(ctx, data.ListID, data.UserID)
	if err != nil {
		return nil, err
	}

	var headingsResp []domain.HeadingResponseData

	for _, heading := range headings {
		headingsResp = append(headingsResp, mapHeadingToResponseData(heading))
	}

	return headingsResp, nil
}

func mapHeadingToResponseData(heading domain.Heading) domain.HeadingResponseData {
	return domain.HeadingResponseData{
		ID:        heading.ID,
		Title:     heading.Title,
		ListID:    heading.ListID,
		UserID:    heading.UserID,
		UpdatedAt: heading.UpdatedAt,
	}
}

func (u *HeadingUsecase) UpdateHeading(ctx context.Context, data *domain.HeadingRequestData) error {
	updatedAt := time.Now().UTC()

	updatedHeading := domain.Heading{
		ID:        data.ID,
		Title:     data.Title,
		ListID:    data.ListID,
		UserID:    data.UserID,
		UpdatedAt: &updatedAt,
	}

	return u.headingStorage.UpdateHeading(ctx, updatedHeading)
}

func (u *HeadingUsecase) MoveHeadingToAnotherList(ctx context.Context, data domain.HeadingRequestData) error {
	updatedAt := time.Now().UTC()

	updatedHeading := domain.Heading{
		ID:        data.ID,
		ListID:    data.ListID,
		UserID:    data.UserID,
		UpdatedAt: &updatedAt,
	}

	updatedTask := domain.Task{
		HeadingID: data.ID,
		ListID:    data.ListID,
		UserID:    data.UserID,
		UpdatedAt: &updatedAt,
	}

	return u.headingStorage.MoveHeadingToAnotherList(ctx, updatedHeading, updatedTask)
}

func (u *HeadingUsecase) DeleteHeading(ctx context.Context, data domain.HeadingRequestData) error {
	deletedAt := time.Now().UTC()

	deletedHeading := domain.Heading{
		ID:        data.ID,
		UserID:    data.UserID,
		DeletedAt: &deletedAt,
	}

	return u.headingStorage.DeleteHeading(ctx, deletedHeading)
}
