package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshelekhov/reframed/internal/domain"
)

type HeadingStorage struct {
	*pgxpool.Pool
}

func NewHeadingStorage(pool *pgxpool.Pool) *HeadingStorage {
	return &HeadingStorage{Pool: pool}
}

func (s *HeadingStorage) CreateHeading(ctx context.Context, heading domain.Heading) error {
	const (
		op = "heading.storage.CreateHeading"

		query = `
			INSERT INTO headings (id, title, list_id, user_id, is_default, updated_at)
			VALUES($1, $2, $3, $4, $5, $6)`
	)

	_, err := s.Exec(
		ctx,
		query,
		heading.ID,
		heading.Title,
		heading.ListID,
		heading.UserID,
		heading.IsDefault,
		heading.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to create heading: %w", op, err)
	}

	return nil
}

func (s *HeadingStorage) GetDefaultHeadingID(ctx context.Context, listID, userID string) (string, error) {
	const (
		op = "heading.storage.GetDefaultHeadingID"

		query = `
			SELECT id
			FROM headings
			WHERE list_id = $1
			  AND user_id = $2
			  AND is_default = true
			  AND deleted_at IS NULL`
	)

	var headingID string

	err := s.QueryRow(ctx, query, listID, userID).Scan(&headingID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrHeadingNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: failed to get default heading: %w", op, err)
	}

	return headingID, nil
}

func (s *HeadingStorage) GetHeadingByID(ctx context.Context, headingID, userID string) (domain.Heading, error) {
	const (
		op = "heading.storage.GetHeadingByID"

		query = `
			SELECT id, title, list_id, user_id, updated_at
			FROM headings
			WHERE id = $1
			  AND user_id = $2
			  AND deleted_at IS NULL`
	)

	var heading domain.Heading

	err := s.QueryRow(ctx, query, headingID, userID).Scan(
		&heading.ID,
		&heading.Title,
		&heading.ListID,
		&heading.UserID,
		&heading.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Heading{}, domain.ErrHeadingNotFound
	}
	if err != nil {
		return domain.Heading{}, fmt.Errorf("%s: failed to get heading: %w", op, err)
	}

	return heading, nil
}

func (s *HeadingStorage) GetHeadingsByListID(ctx context.Context, listID, userID string) ([]domain.Heading, error) {
	const (
		op = "heading.storage.GetHeadingsByListID"

		query = `
			SELECT id, title, list_id, user_id, updated_at
			FROM headings
			WHERE list_id = $1
			  AND user_id = $2
			  AND deleted_at IS NULL
		`
	)

	rows, err := s.Query(ctx, query, listID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	defer rows.Close()

	var headings []domain.Heading

	for rows.Next() {
		heading := domain.Heading{}

		err = rows.Scan(
			&heading.ID,
			&heading.Title,
			&heading.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan heading: %w", op, err)
		}

		headings = append(headings, heading)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to iterate over rows: %w", op, err)
	}

	if len(headings) == 0 {
		return nil, domain.ErrNoHeadingsFound
	}

	return headings, nil
}

func (s *HeadingStorage) UpdateHeading(ctx context.Context, heading domain.Heading) error {
	const (
		op = "heading.storage.UpdateHeading"

		query = `
			UPDATE headings
			SET title = $1, updated_at = $2
			WHERE id = $3
			  AND user_id = $4`
	)

	result, err := s.Exec(
		ctx,
		query,
		heading.Title,
		heading.UpdatedAt,
		heading.ID,
		heading.UserID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to update heading: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrHeadingNotFound
	}

	return nil
}

func (s *HeadingStorage) MoveHeadingToAnotherList(ctx context.Context, heading domain.Heading, task domain.Task) error {
	const (
		op = "heading.storage.MoveTaskToAnotherList"

		queryUpdateHeading = `
			UPDATE headings
			SET list_id = $1, updated_at = $2
			WHERE id = $3
			  AND user_id = $4`

		queryUpdateTasks = `
			UPDATE tasks
			SET list_id = $1, updated_at = $2
			WHERE heading_id = $3
			  AND user_id = $4`
	)

	updateHeading, err := s.Exec(
		ctx,
		queryUpdateHeading,
		heading.ListID,
		heading.UpdatedAt,
		heading.ID,
		heading.UserID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to update heading: %w", op, err)
	}

	if updateHeading.RowsAffected() == 0 {
		return domain.ErrHeadingNotFound
	}

	updateTasks, err := s.Exec(
		ctx,
		queryUpdateTasks,
		task.ListID,
		task.UpdatedAt,
		task.HeadingID,
		task.UserID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to update tasks: %w", op, err)
	}

	if updateTasks.RowsAffected() == 0 {
		// It means that this heading has no tasks
		return nil
	}

	return nil
}

func (s *HeadingStorage) DeleteHeading(ctx context.Context, heading domain.Heading) error {
	const (
		op = "heading.storage.DeleteHeading"

		query = `
			UPDATE headings
			SET deleted_at = $1
			WHERE id = $2
			  AND user_id = $3
			  AND deleted_at IS NULL`
	)

	result, err := s.Exec(
		ctx,
		query,
		heading.DeletedAt,
		heading.ID,
		heading.UserID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to delete heading: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrHeadingNotFound
	}

	return nil
}
