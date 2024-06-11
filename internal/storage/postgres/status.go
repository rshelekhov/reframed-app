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

type StatusStorage struct {
	*pgxpool.Pool
	*sqlc.Queries
}

func NewStatusStorage(pool *pgxpool.Pool) *StatusStorage {
	return &StatusStorage{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}

func (s *StatusStorage) GetStatuses(ctx context.Context) ([]model.Status, error) {
	const op = "status.storage.GetStatuses"

	items, err := s.Queries.GetStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tags: %w", op, err)
	}
	if len(items) == 0 {
		return nil, le.ErrNoStatusesFound
	}

	var statuses []model.Status

	for _, item := range items {
		statuses = append(statuses, model.Status{
			ID:    item.ID,
			Title: item.Title,
		})
	}

	return statuses, nil
}

func (s *StatusStorage) GetStatusByID(ctx context.Context, statusID int32) (string, error) {
	const op = "status.storage.GetStatusName"

	statusName, err := s.Queries.GetStatusByID(ctx, statusID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", le.ErrStatusNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: failed to get statusName: %w", op, err)
	}

	return statusName, nil
}
