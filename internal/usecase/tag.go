package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type TagUsecase struct {
	storage port.TagStorage
}

func NewTagUsecase(storage port.TagStorage) *TagUsecase {
	return &TagUsecase{storage: storage}
}

func (u *TagUsecase) CreateTagIfNotExists(ctx context.Context, data model.TagRequestData) error {
	_, err := u.storage.GetTagIDByTitle(ctx, data.Title, data.UserID)
	if errors.Is(err, le.ErrTagNotFound) {
		newTag := model.Tag{
			ID:        ksuid.New().String(),
			Title:     data.Title,
			UserID:    data.UserID,
			UpdatedAt: time.Now(),
		}

		return u.storage.CreateTag(ctx, newTag)
	}
	if err != nil {
		return err
	}

	return nil
}

func (u *TagUsecase) LinkTagsToTask(ctx context.Context, userID, taskID string, tags []string) error {
	return u.storage.LinkTagsToTask(ctx, userID, taskID, tags)
}

func (u *TagUsecase) UnlinkTagsFromTask(ctx context.Context, userID, taskID string, tags []string) error {
	return u.storage.UnlinkTagsFromTask(ctx, userID, taskID, tags)
}

func (u *TagUsecase) GetTagsByUserID(ctx context.Context, userID string) ([]model.TagResponseData, error) {
	tags, err := u.storage.GetTagsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var tagsResp []model.TagResponseData

	for _, tag := range tags {
		tagsResp = append(tagsResp, mapTagToTagResponseData(tag))
	}

	return tagsResp, nil
}

func (u *TagUsecase) GetTagsByTaskID(ctx context.Context, taskID string) ([]model.TagResponseData, error) {
	tags, err := u.storage.GetTagsByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	var tagsResp []model.TagResponseData

	for _, tag := range tags {
		tagsResp = append(tagsResp, mapTagToTagResponseData(tag))
	}

	return tagsResp, nil
}

func mapTagToTagResponseData(tag model.Tag) model.TagResponseData {
	return model.TagResponseData{
		ID:        tag.ID,
		Title:     tag.Title,
		UpdatedAt: tag.UpdatedAt,
	}
}
