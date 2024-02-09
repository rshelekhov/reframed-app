package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/models"
	"github.com/segmentio/ksuid"
	"time"
)

type TagStorage struct {
	*pgxpool.Pool
}

func NewTagStorage(pool *pgxpool.Pool) *TagStorage {
	return &TagStorage{Pool: pool}
}

func (s *TagStorage) CreateTagIfNotExists(ctx context.Context, tag, userID string) error {
	const (
		op = "tag.storage.CreateTagIfNotExists"

		querySelectTag = `
			SELECT id
			FROM tags
			WHERE title = $1
			  AND user_id = $2
			  AND deleted_at IS NULL`

		queryInsertTag = `
			INSERT INTO tags (id, title, user_id, updated_at)
			VALUES ($1, LOWER($2), $3, $4)`
	)

	// Check if tag exists
	var count int
	err := s.QueryRow(ctx, querySelectTag, tag, userID).Scan(&count)
	if err != nil {
		return fmt.Errorf("%s: failed to check if tag existsy: %w", op, err)
	}

	// Insert tag if not exists
	if count == 0 {
		_, err = s.Exec(
			ctx,
			queryInsertTag,
			ksuid.New().String(),
			tag,
			userID,
			time.Now().UTC(),
		)
		if err != nil {
			return fmt.Errorf("%s: failed to insert tag: %w", op, err)
		}
	}

	return nil
}

func (s *TagStorage) LinkTagsToTask(ctx context.Context, taskID string, tags []string) error {
	const (
		op = "tag.storage.LinkTagsToTask"

		query = `
			INSERT INTO tasks_tags (task_id, tag_id)
			VALUES ($1, (SELECT id
			      			FROM tags
			      			WHERE title = lower($2))
			)`
	)

	for _, tag := range tags {
		_, err := s.Exec(ctx, query, taskID, tag)
		if err != nil {
			return fmt.Errorf("%s: failed to link tag to task: %w", op, err)
		}
	}

	return nil
}

func (s *TagStorage) UnlinkTagsFromTask(ctx context.Context, taskID string, tags []string) error {
	const (
		op = "tag.storage.UnlinkTagsFromTask"

		query = `
			DELETE FROM tasks_tags
			WHERE task_id = $1 AND tag_id =
			(
				SELECT id
				FROM tags
				WHERE title = lower($2)
			)`
	)

	for _, tag := range tags {
		_, err := s.Exec(ctx, query, taskID, tag)
		if err != nil {
			return fmt.Errorf("%s: failed to unlink tag from task: %w", op, err)
		}
	}

	return nil
}

func (s *TagStorage) GetTagsByUserID(ctx context.Context, userID string) ([]models.Tag, error) {
	const (
		op = "tag.storage.GetTagsByUserID"

		query = `
			SELECT id, title, updated_at
			FROM tags
			WHERE user_id = $1
			  AND deleted_at IS NULL`
	)

	rows, err := s.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	defer rows.Close()

	var tags []models.Tag

	for rows.Next() {
		tag := models.Tag{}

		err = rows.Scan(
			&tag.ID,
			&tag.Title,
			&tag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan tag: %w", op, err)
		}

		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(tags) == 0 {
		return nil, c.ErrNoTagsFound
	}

	return tags, nil
}

func (s *TagStorage) GetTagsByTaskID(ctx context.Context, taskID string) ([]string, error) {
	const (
		op = "tag.storage.GetTagsByTaskID"

		query = `
			SELECT tags.title
			FROM tags
				JOIN tasks_tags
				    ON tags.id = tasks_tags.tag_id
			WHERE tasks_tags.task_id = $1
			  AND tags.deleted_at IS NULL`
	)

	rows, err := s.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	defer rows.Close()

	var tags []string

	for rows.Next() {
		var tag string

		err = rows.Scan(&tag)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan tag: %w", op, err)
		}

		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(tags) == 0 {
		return nil, c.ErrNoTagsFound
	}

	return tags, nil
}
