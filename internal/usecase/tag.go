package usecase

import (
	"context"
	"errors"
	"github.com/rshelekhov/reframed/internal/domain"
	"github.com/segmentio/ksuid"
	"time"
)

type TagUsecase struct {
	tagStorage domain.TagStorage
}

func NewTagUsecase(storage domain.TagStorage) *TagUsecase {
	return &TagUsecase{
		tagStorage: storage,
	}
}

func (u *TagUsecase) CreateTagIfNotExists(ctx context.Context, data domain.TagRequestData) error {
	_, err := u.tagStorage.GetTagIDByTitle(ctx, data.Title, data.UserID)
	if errors.Is(err, domain.ErrTagNotFound) {
		// Create new tag
		updatedAt := time.Now().UTC()

		newTag := domain.Tag{
			ID:        ksuid.New().String(),
			Title:     data.Title,
			UserID:    data.UserID,
			UpdatedAt: &updatedAt,
		}
		return u.tagStorage.CreateTag(ctx, newTag)
	}
	if err != nil {
		return err
	}

	return nil
}

func (u *TagUsecase) LinkTagsToTask(ctx context.Context, taskID string, tags []string) error {
	return u.tagStorage.LinkTagsToTask(ctx, taskID, tags)
}

func (u *TagUsecase) UnlinkTagsFromTask(ctx context.Context, taskID string, tagsToRemove []string) error {
	return u.tagStorage.UnlinkTagsFromTask(ctx, taskID, tagsToRemove)
}

func (u *TagUsecase) GetTagsByUserID(ctx context.Context, userID string) ([]domain.TagResponseData, error) {
	tags, err := u.tagStorage.GetTagsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var tagsResp []domain.TagResponseData

	for _, tag := range tags {
		tagsResp = append(tagsResp, mapTagToTagResponseData(tag))
	}

	return tagsResp, nil
}

func (u *TagUsecase) GetTagsByTaskID(ctx context.Context, taskID string) ([]domain.TagResponseData, error) {
	tags, err := u.tagStorage.GetTagsByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	var tagsResp []domain.TagResponseData

	for _, tag := range tags {
		tagsResp = append(tagsResp, mapTagToTagResponseData(tag))
	}

	return tagsResp, nil
}

func mapTagToTagResponseData(tag domain.Tag) domain.TagResponseData {
	return domain.TagResponseData{
		ID:        tag.ID,
		Title:     tag.Title,
		UpdatedAt: tag.UpdatedAt,
	}
}
