package domain

import (
	"context"
	"time"
)

// Tag DB domain
type (
	Tag struct {
		ID        string     `db:"id" json:"id,omitempty"`
		Title     string     `db:"headingTitle" json:"headingTitle,omitempty"`
		UserID    string     `db:"user_id" json:"user_id,omitempty"`
		UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
		DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	}

	TagRequestData struct {
		Title  string `json:"title" validate:"required"`
		UserID string `json:"user_id"`
	}

	TagResponseData struct {
		ID        string     `json:"id,omitempty"`
		Title     string     `json:"title,omitempty"`
		UpdatedAt *time.Time `json:"updated_at,omitempty"`
	}
)

type (
	TagUsecase interface {
		CreateTagIfNotExists(ctx context.Context, data TagRequestData) error
		LinkTagsToTask(ctx context.Context, taskID string, tags []string) error
		UnlinkTagsFromTask(ctx context.Context, taskID string, tagsToRemove []string) error
		GetTagsByUserID(ctx context.Context, userID string) ([]TagResponseData, error)
		GetTagsByTaskID(ctx context.Context, taskID string) ([]TagResponseData, error)
	}

	TagStorage interface {
		CreateTag(ctx context.Context, tag Tag) error
		LinkTagsToTask(ctx context.Context, taskID string, tags []string) error
		UnlinkTagsFromTask(ctx context.Context, taskID string, tagsToRemove []string) error
		GetTagIDByTitle(ctx context.Context, title, userID string) (string, error)
		GetTagsByUserID(ctx context.Context, userID string) ([]Tag, error)
		GetTagsByTaskID(ctx context.Context, taskID string) ([]Tag, error)
	}
)
