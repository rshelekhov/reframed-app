package usecase

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
	"github.com/segmentio/ksuid"
	"time"
)

type HeadingUsecase struct {
	headingStorage port.HeadingStorage
}

func NewHeadingUsecase(storage port.HeadingStorage) *HeadingUsecase {
	return &HeadingUsecase{
		headingStorage: storage,
	}
}

func (u *HeadingUsecase) CreateHeading(ctx context.Context, data *model.HeadingRequestData) (model.HeadingResponseData, error) {
	newHeading := model.Heading{
		ID:        ksuid.New().String(),
		Title:     data.Title,
		ListID:    data.ListID,
		UserID:    data.UserID,
		IsDefault: false,
		UpdatedAt: time.Now(),
	}

	if err := u.headingStorage.CreateHeading(ctx, newHeading); err != nil {
		return model.HeadingResponseData{}, err
	}

	return model.HeadingResponseData{
		ID:        newHeading.ID,
		Title:     newHeading.Title,
		ListID:    newHeading.ListID,
		UserID:    newHeading.UserID,
		UpdatedAt: newHeading.UpdatedAt,
	}, nil
}

func (u *HeadingUsecase) CreateDefaultHeading(ctx context.Context, heading model.Heading) error {
	return u.headingStorage.CreateHeading(ctx, heading)
}

func (u *HeadingUsecase) GetHeadingByID(ctx context.Context, data model.HeadingRequestData) (model.HeadingResponseData, error) {
	heading, err := u.headingStorage.GetHeadingByID(ctx, data.ID, data.UserID)
	if err != nil {
		return model.HeadingResponseData{}, err
	}

	return model.HeadingResponseData{
		ID:        heading.ID,
		Title:     heading.Title,
		ListID:    heading.ListID,
		UserID:    heading.UserID,
		UpdatedAt: heading.UpdatedAt,
	}, nil
}

func (u *HeadingUsecase) GetDefaultHeadingID(ctx context.Context, data model.HeadingRequestData) (string, error) {
	return u.headingStorage.GetDefaultHeadingID(ctx, data.ListID, data.UserID)
}

func (u *HeadingUsecase) GetHeadingsByListID(ctx context.Context, data model.HeadingRequestData) ([]model.HeadingResponseData, error) {
	headings, err := u.headingStorage.GetHeadingsByListID(ctx, data.ListID, data.UserID)
	if err != nil {
		return nil, err
	}

	var headingsResp []model.HeadingResponseData

	for _, heading := range headings {
		headingsResp = append(headingsResp, mapHeadingToResponseData(heading))
	}

	return headingsResp, nil
}

func mapHeadingToResponseData(heading model.Heading) model.HeadingResponseData {
	return model.HeadingResponseData{
		ID:        heading.ID,
		Title:     heading.Title,
		ListID:    heading.ListID,
		UserID:    heading.UserID,
		UpdatedAt: heading.UpdatedAt,
	}
}

func (u *HeadingUsecase) UpdateHeading(ctx context.Context, data *model.HeadingRequestData) (model.HeadingResponseData, error) {
	updatedHeading := model.Heading{
		ID:        data.ID,
		Title:     data.Title,
		ListID:    data.ListID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	if err := u.headingStorage.UpdateHeading(ctx, updatedHeading); err != nil {
		return model.HeadingResponseData{}, err
	}

	return model.HeadingResponseData{
		ID:        updatedHeading.ID,
		Title:     updatedHeading.Title,
		ListID:    updatedHeading.ListID,
		UserID:    updatedHeading.UserID,
		UpdatedAt: updatedHeading.UpdatedAt,
	}, nil
}

func (u *HeadingUsecase) MoveHeadingToAnotherList(ctx context.Context, data model.HeadingRequestData) (model.HeadingResponseData, error) {
	updatedHeading := model.Heading{
		ID:        data.ID,
		ListID:    data.ListID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	updatedTasks := model.Task{
		HeadingID: data.ID,
		ListID:    data.ListID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	if err := u.headingStorage.MoveHeadingToAnotherList(ctx, updatedHeading, updatedTasks); err != nil {
		return model.HeadingResponseData{}, err
	}

	return model.HeadingResponseData{
		ID:        updatedHeading.ID,
		Title:     updatedHeading.Title,
		ListID:    updatedHeading.ListID,
		UserID:    updatedHeading.UserID,
		UpdatedAt: updatedHeading.UpdatedAt,
	}, nil
}

func (u *HeadingUsecase) DeleteHeading(ctx context.Context, data model.HeadingRequestData) error {
	deletedHeading := model.Heading{
		ID:        data.ID,
		UserID:    data.UserID,
		DeletedAt: time.Now(),
	}

	return u.headingStorage.DeleteHeading(ctx, deletedHeading)
}
