package port

import (
	"context"

	"github.com/rshelekhov/reframed/internal/model"
)

type (
	TagUsecase interface {
		CreateTagIfNotExists(ctx context.Context, data model.TagRequestData) error
		LinkTagsToTask(ctx context.Context, taskID string, tags []string) error
		UnlinkTagsFromTask(ctx context.Context, taskID string, tagsToRemove []string) error
		GetTagsByUserID(ctx context.Context, userID string) ([]model.TagResponseData, error)
		GetTagsByTaskID(ctx context.Context, taskID string) ([]model.TagResponseData, error)
	}

	TagStorage interface {
		CreateTag(ctx context.Context, tag model.Tag) error
		LinkTagsToTask(ctx context.Context, taskID string, tags []string) error
		UnlinkTagsFromTask(ctx context.Context, taskID string, tagsToRemove []string) error
		GetTagIDByTitle(ctx context.Context, title, userID string) (string, error)
		GetTagsByUserID(ctx context.Context, userID string) ([]model.Tag, error)
		GetTagsByTaskID(ctx context.Context, taskID string) ([]model.Tag, error)
	}
)
