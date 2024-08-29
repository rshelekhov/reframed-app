package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/storage/postgres/sqlc"
)

type HeadingStorage struct {
	*pgxpool.Pool
	*sqlc.Queries
}

func NewHeadingStorage(pool *pgxpool.Pool) *HeadingStorage {
	return &HeadingStorage{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}

func (s *HeadingStorage) CreateHeading(ctx context.Context, heading model.Heading) error {
	const op = "heading.storage.CreateHeading"

	if err := s.Queries.CreateHeading(ctx, sqlc.CreateHeadingParams{
		ID:        heading.ID,
		Title:     heading.Title,
		ListID:    heading.ListID,
		UserID:    heading.UserID,
		IsDefault: heading.IsDefault,
		CreatedAt: heading.CreatedAt,
		UpdatedAt: heading.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("%s: failed to create heading: %w", op, err)
	}
	return nil
}

func (s *HeadingStorage) GetDefaultHeadingID(ctx context.Context, listID, userID string) (string, error) {
	const op = "heading.storage.GetDefaultHeadingID"

	headingID, err := s.Queries.GetDefaultHeadingID(ctx, sqlc.GetDefaultHeadingIDParams{
		ListID: listID,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", le.ErrDefaultHeadingNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: failed to get default heading: %w", op, err)
	}

	return headingID, nil
}

func (s *HeadingStorage) GetHeadingByID(ctx context.Context, headingID, userID string) (model.Heading, error) {
	const op = "heading.storage.GetHeadingByID"

	heading, err := s.Queries.GetHeadingByID(ctx, sqlc.GetHeadingByIDParams{
		ID:     headingID,
		UserID: userID,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Heading{}, le.ErrHeadingNotFound
	}
	if err != nil {
		return model.Heading{}, fmt.Errorf("%s: failed to get heading: %w", op, err)
	}

	return model.Heading{
		ID:        heading.ID,
		Title:     heading.Title,
		ListID:    heading.ListID,
		UserID:    heading.UserID,
		UpdatedAt: heading.UpdatedAt,
	}, nil
}

func (s *HeadingStorage) GetHeadingsByListID(ctx context.Context, listID, userID string) ([]model.Heading, error) {
	const op = "heading.storage.GetHeadingsByListID"

	items, err := s.Queries.GetHeadingsByListID(ctx, sqlc.GetHeadingsByListIDParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get headings: %w", op, err)
	}
	if len(items) == 0 {
		return nil, le.ErrNoHeadingsFound
	}

	var headings []model.Heading

	for _, item := range items {
		headings = append(headings, model.Heading{
			ID:        item.ID,
			Title:     item.Title,
			ListID:    item.ListID,
			UserID:    item.UserID,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return headings, nil
}

func (s *HeadingStorage) UpdateHeading(ctx context.Context, heading model.Heading) error {
	const op = "heading.storage.UpdateHeading"

	_, err := s.Queries.UpdateHeading(ctx, sqlc.UpdateHeadingParams{
		Title:     heading.Title,
		UpdatedAt: heading.UpdatedAt,
		ID:        heading.ID,
		UserID:    heading.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrHeadingNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to update heading: %w", op, err)
	}

	return nil
}

func (s *HeadingStorage) MoveHeadingToAnotherList(ctx context.Context, heading model.Heading, task model.Task) error {
	const op = "heading.storage.MoveTaskToAnotherList"

	_, err := s.Queries.MoveHeadingToAnotherList(ctx, sqlc.MoveHeadingToAnotherListParams{
		ListID:    heading.ListID,
		UpdatedAt: heading.UpdatedAt,
		ID:        heading.ID,
		UserID:    heading.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrHeadingNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to update heading: %w", op, err)
	}

	err = s.Queries.UpdateTasksListID(ctx, sqlc.UpdateTasksListIDParams{
		ListID:    task.ListID,
		UpdatedAt: task.UpdatedAt,
		HeadingID: task.HeadingID,
		UserID:    task.UserID,
	})
	if err != nil {
		return fmt.Errorf("%s: failed to update tasks: %w", op, err)
	}

	return nil
}

func (s *HeadingStorage) DeleteHeading(ctx context.Context, heading model.Heading) error {
	const op = "heading.storage.DeleteHeading"

	_, err := s.Queries.DeleteHeading(ctx, sqlc.DeleteHeadingParams{
		ID:     heading.ID,
		UserID: heading.UserID,
		DeletedAt: pgtype.Timestamptz{
			Time:  heading.DeletedAt,
			Valid: true,
		},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrHeadingNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to delete heading: %w", op, err)
	}

	return nil
}
