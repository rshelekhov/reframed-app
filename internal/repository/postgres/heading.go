package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/pkg/constants/le"
)

type HeadingStorage struct {
	*pgxpool.Pool
	*Queries
}

func NewHeadingStorage(pool *pgxpool.Pool) *HeadingStorage {
	return &HeadingStorage{
		Pool:    pool,
		Queries: New(pool),
	}
}

func (s *HeadingStorage) CreateHeading(ctx context.Context, heading model.Heading) error {
	const op = "heading.repository.CreateHeading"

	if err := s.Queries.CreateHeading(ctx, CreateHeadingParams{
		ID:        heading.ID,
		Title:     heading.Title,
		ListID:    heading.ListID,
		UserID:    heading.UserID,
		IsDefault: heading.IsDefault,
		UpdatedAt: pgtype.Timestamptz{
			Time: heading.UpdatedAt,
		},
	}); err != nil {
		return fmt.Errorf("%s: failed to create heading: %w", op, err)
	}
	return nil
}

func (s *HeadingStorage) GetDefaultHeadingID(ctx context.Context, listID, userID string) (string, error) {
	const op = "heading.repository.GetDefaultHeadingID"

	headingID, err := s.Queries.GetDefaultHeadingID(ctx, GetDefaultHeadingIDParams{
		ListID: listID,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", le.ErrHeadingNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: failed to get default heading: %w", op, err)
	}

	return headingID, nil
}

func (s *HeadingStorage) GetHeadingByID(ctx context.Context, headingID, userID string) (model.Heading, error) {
	const op = "heading.repository.GetHeadingByID"

	heading, err := s.Queries.GetHeadingByID(ctx, GetHeadingByIDParams{
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
		UpdatedAt: heading.UpdatedAt.Time,
	}, nil
}

func (s *HeadingStorage) GetHeadingsByListID(ctx context.Context, listID, userID string) ([]model.Heading, error) {
	const op = "heading.repository.GetHeadingsByListID"

	items, err := s.Queries.GetHeadingsByListID(ctx, GetHeadingsByListIDParams{
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
			UpdatedAt: item.UpdatedAt.Time,
		})
	}
	return headings, nil
}

func (s *HeadingStorage) UpdateHeading(ctx context.Context, heading model.Heading) error {
	const op = "heading.repository.UpdateHeading"

	err := s.Queries.UpdateHeading(ctx, UpdateHeadingParams{
		Title: heading.Title,
		UpdatedAt: pgtype.Timestamptz{
			Time: heading.UpdatedAt,
		},
		ID:     heading.ID,
		UserID: heading.UserID,
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
	const op = "heading.repository.MoveTaskToAnotherList"

	err := s.Queries.MoveHeadingToAnotherList(ctx, MoveHeadingToAnotherListParams{
		ListID: heading.ListID,
		UpdatedAt: pgtype.Timestamptz{
			Time: heading.UpdatedAt,
		},
		ID:     heading.ID,
		UserID: heading.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrHeadingNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to update heading: %w", op, err)
	}

	err = s.Queries.UpdateTasksListID(ctx, UpdateTasksListIDParams{
		ListID: task.ListID,
		UpdatedAt: pgtype.Timestamptz{
			Time: task.UpdatedAt,
		},
		HeadingID: pgtype.Text{
			String: task.HeadingID,
		},
		UserID: task.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrHeadingNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to update tasks: %w", op, err)
	}

	return nil
}

func (s *HeadingStorage) DeleteHeading(ctx context.Context, heading model.Heading) error {
	const op = "heading.repository.DeleteHeading"

	err := s.Queries.DeleteHeading(ctx, DeleteHeadingParams{
		ID:     heading.ID,
		UserID: heading.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrHeadingNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to delete heading: %w", op, err)
	}

	return nil
}
