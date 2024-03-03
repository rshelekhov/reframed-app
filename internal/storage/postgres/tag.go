package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/pkg/constants/le"
)

type TagStorage struct {
	*pgxpool.Pool
	*Queries
}

func NewTagStorage(pool *pgxpool.Pool) *TagStorage {
	return &TagStorage{
		Pool:    pool,
		Queries: New(pool),
	}
}

func (s *TagStorage) CreateTag(ctx context.Context, tag model.Tag) error {
	const op = "tag.storage.CreateTag"

	if err := s.Queries.CreateTag(ctx, CreateTagParams{
		ID:        tag.ID,
		Title:     tag.Title,
		UserID:    tag.UserID,
		UpdatedAt: tag.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("%s: failed to insert tag: %w", op, err)
	}
	return nil
}

func (s *TagStorage) LinkTagsToTask(ctx context.Context, taskID string, tags []string) error {
	const op = "tag.storage.LinkTagsToTask"

	for _, tag := range tags {
		if err := s.Queries.LinkTagToTask(ctx, LinkTagToTaskParams{
			TaskID: taskID,
			Title:  tag,
		}); err != nil {
			return fmt.Errorf("%s: failed to link tag to task: %w", op, err)
		}
	}
	return nil
}

func (s *TagStorage) UnlinkTagsFromTask(ctx context.Context, taskID string, tags []string) error {
	const op = "tag.storage.UnlinkTagsFromTask"

	for _, tag := range tags {
		if err := s.Queries.UnlinkTagFromTask(ctx, UnlinkTagFromTaskParams{
			TaskID: taskID,
			Title:  tag,
		}); err != nil {
			return fmt.Errorf("%s: failed to unlink tag from task: %w", op, err)
		}
	}
	return nil
}

func (s *TagStorage) GetTagIDByTitle(ctx context.Context, title, userID string) (string, error) {
	const op = "tag.storage.GetTagByTitle"

	tagID, err := s.Queries.GetTagIDByTitle(ctx, GetTagIDByTitleParams{
		Title:  title,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", le.ErrTagNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: failed to check if tag exists: %w", op, err)
	}

	return tagID, nil
}

func (s *TagStorage) GetTagsByUserID(ctx context.Context, userID string) ([]model.Tag, error) {
	const op = "tag.storage.GetTagsByUserID"

	items, err := s.Queries.GetTagsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tags: %w", op, err)
	}
	if len(items) == 0 {
		return nil, le.ErrNoTagsFound
	}

	var tags []model.Tag

	for _, item := range items {
		tags = append(tags, model.Tag{
			ID:        item.ID,
			Title:     item.Title,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return tags, nil
}

func (s *TagStorage) GetTagsByTaskID(ctx context.Context, taskID string) ([]model.Tag, error) {
	const op = "tag.storage.GetTagsByTaskID"

	var tagsTitles []model.Tag

	tags, err := s.Queries.GetTagsByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tags: %w", op, err)
	}
	if len(tagsTitles) == 0 {
		return nil, le.ErrNoTagsFound
	}

	for _, tag := range tags {
		tagsTitles = append(tagsTitles, model.Tag{
			ID:        tag.ID,
			Title:     tag.Title,
			UpdatedAt: tag.UpdatedAt,
		})
	}
	return tagsTitles, nil
}
