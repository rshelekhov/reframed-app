package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/storage/postgres/sqlc"
)

type TagStorage struct {
	*pgxpool.Pool
	*sqlc.Queries
}

func NewTagStorage(pool *pgxpool.Pool) *TagStorage {
	return &TagStorage{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}

func (s *TagStorage) CreateTag(ctx context.Context, tag model.Tag) error {
	const op = "tag.storage.CreateTag"

	if err := s.Queries.CreateTag(ctx, sqlc.CreateTagParams{
		ID:        tag.ID,
		Title:     tag.Title,
		UserID:    tag.UserID,
		UpdatedAt: tag.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("%s: failed to insert tag: %w", op, err)
	}
	return nil
}

func (s *TagStorage) LinkTagsToTask(ctx context.Context, userID, taskID string, tags []string) error {
	const op = "tag.storage.LinkTagsToTask"

	for _, tag := range tags {
		if err := s.Queries.LinkTagToTask(ctx, sqlc.LinkTagToTaskParams{
			TaskID: taskID,
			Title:  tag,
			UserID: userID,
		}); err != nil {
			return fmt.Errorf("%s: failed to link tag to task: %w", op, err)
		}
	}
	return nil
}

func (s *TagStorage) UnlinkTagsFromTask(ctx context.Context, userID, taskID string, tags []string) error {
	const op = "tag.storage.UnlinkTagsFromTask"

	for _, tag := range tags {
		if err := s.Queries.UnlinkTagFromTask(ctx, sqlc.UnlinkTagFromTaskParams{
			TaskID: taskID,
			Title:  tag,
			UserID: userID,
		}); err != nil {
			return fmt.Errorf("%s: failed to unlink tag from task: %w", op, err)
		}
	}
	return nil
}

func (s *TagStorage) GetTagIDByTitle(ctx context.Context, title, userID string) (string, error) {
	const op = "tag.storage.GetTagByTitle"

	tagID, err := s.Queries.GetTagIDByTitle(ctx, sqlc.GetTagIDByTitleParams{
		Title:  title,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", le.ErrTagNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: failed to get tag_id by title: %w", op, err)
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
		// There are no tags, it's ok
		return nil, nil
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
