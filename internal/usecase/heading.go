package usecase

import (
	"context"
	"time"

	"github.com/rshelekhov/reframed/internal/lib/constant/le"

	"github.com/segmentio/ksuid"

	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type HeadingUsecase struct {
	storage     port.HeadingStorage
	ListUsecase port.ListUsecase
	TaskUsecase port.TaskUsecase
}

func NewHeadingUsecase(storage port.HeadingStorage) *HeadingUsecase {
	return &HeadingUsecase{storage: storage}
}

func (u *HeadingUsecase) CreateHeading(ctx context.Context, data *model.HeadingRequestData) (model.HeadingResponseData, error) {
	err := u.handleListID(ctx, data)
	if err != nil {
		return model.HeadingResponseData{}, err
	}

	currentTime := time.Now()

	newHeading := model.Heading{
		ID:        ksuid.New().String(),
		Title:     data.Title,
		ListID:    data.ListID,
		UserID:    data.UserID,
		IsDefault: false,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	if err = u.storage.CreateHeading(ctx, newHeading); err != nil {
		return model.HeadingResponseData{}, err
	}

	return model.HeadingResponseData{
		ID:        newHeading.ID,
		Title:     newHeading.Title,
		ListID:    newHeading.ListID,
		UserID:    newHeading.UserID,
		CreatedAt: newHeading.CreatedAt,
		UpdatedAt: newHeading.UpdatedAt,
	}, nil
}

func (u *HeadingUsecase) handleListID(ctx context.Context, data *model.HeadingRequestData) error {
	// Check if list exists
	list, err := u.ListUsecase.GetListByID(ctx, model.ListRequestData{ID: data.ListID, UserID: data.UserID})
	if err != nil {
		return err
	}

	// Check that this list belongs to the user
	if list.UserID != data.UserID {
		return le.ErrListNotFound
	}

	return nil
}

func (u *HeadingUsecase) CreateDefaultHeading(ctx context.Context, heading model.Heading) error {
	return u.storage.CreateHeading(ctx, heading)
}

func (u *HeadingUsecase) GetHeadingByID(ctx context.Context, data model.HeadingRequestData) (model.HeadingResponseData, error) {
	heading, err := u.storage.GetHeadingByID(ctx, data.ID, data.UserID)
	if err != nil {
		return model.HeadingResponseData{}, err
	}

	return model.HeadingResponseData{
		ID:        heading.ID,
		Title:     heading.Title,
		ListID:    heading.ListID,
		UserID:    heading.UserID,
		CreatedAt: heading.CreatedAt,
		UpdatedAt: heading.UpdatedAt,
	}, nil
}

func (u *HeadingUsecase) GetDefaultHeadingID(ctx context.Context, data model.HeadingRequestData) (string, error) {
	return u.storage.GetDefaultHeadingID(ctx, data.ListID, data.UserID)
}

func (u *HeadingUsecase) GetHeadingsByListID(ctx context.Context, data model.HeadingRequestData) ([]model.HeadingResponseData, error) {
	headings, err := u.storage.GetHeadingsByListID(ctx, data.ListID, data.UserID)
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
		CreatedAt: heading.CreatedAt,
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

	if err := u.storage.UpdateHeading(ctx, updatedHeading); err != nil {
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

func (u *HeadingUsecase) MoveHeadingToAnotherList(ctx context.Context, data *model.HeadingRequestData) (model.HeadingResponseData, error) {
	err := u.handleListID(ctx, data)
	if err != nil {
		return model.HeadingResponseData{}, err
	}

	currentTime := time.Now()

	updatedHeading := model.Heading{
		ID:        data.ID,
		ListID:    data.ListID,
		UserID:    data.UserID,
		UpdatedAt: currentTime,
	}

	updatedTasks := model.Task{
		HeadingID: data.ID,
		ListID:    data.ListID,
		UserID:    data.UserID,
		UpdatedAt: currentTime,
	}

	if err = u.storage.MoveHeadingToAnotherList(ctx, updatedHeading, updatedTasks); err != nil {
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

func (u *HeadingUsecase) DeleteHeading(ctx context.Context, data *model.HeadingRequestData) error {
	deletedHeading := model.Heading{
		ID:        data.ID,
		UserID:    data.UserID,
		DeletedAt: time.Now(),
	}

	if err := u.storage.DeleteHeading(ctx, deletedHeading); err != nil {
		return err
	}

	tasksData := model.TaskRequestData{
		HeadingID: data.ID,
		UserID:    data.UserID,
	}

	if err := u.TaskUsecase.ArchiveTasksByHeadingID(ctx, tasksData); err != nil {
		return err
	}

	return nil
}

func (u *HeadingUsecase) DeleteHeadingsByListID(ctx context.Context, data model.HeadingRequestData) error {
	deletedHeadings := model.Heading{
		UserID:    data.UserID,
		ListID:    data.ListID,
		DeletedAt: time.Now(),
	}

	if err := u.storage.DeleteHeadingsByListID(ctx, deletedHeadings); err != nil {
		return err
	}

	return nil
}
