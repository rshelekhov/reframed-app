package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/storage/postgres/sqlc"
)

type ListStorage struct {
	*pgxpool.Pool
	*sqlc.Queries
}

func NewListStorage(pool *pgxpool.Pool) *ListStorage {
	return &ListStorage{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}

func (s *ListStorage) CreateList(ctx context.Context, list model.List) error {
	const op = "list.storage.CreateList"

	if err := s.Queries.CreateList(ctx, sqlc.CreateListParams{
		ID:        list.ID,
		Title:     list.Title,
		IsDefault: list.IsDefault,
		UserID:    list.UserID,
		UpdatedAt: list.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("%s: failed to create new list: %w", op, err)
	}
	return nil
}

func (s *ListStorage) GetListByID(ctx context.Context, listID, userID string) (model.List, error) {
	const op = "list.storage.GetListByID"

	list, err := s.Queries.GetListByID(ctx, sqlc.GetListByIDParams{
		ID:     listID,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.List{}, le.ErrListNotFound
	}
	if err != nil {
		return model.List{}, fmt.Errorf("%s: failed to get list: %w", op, err)
	}

	return model.List{
		ID:        list.ID,
		Title:     list.Title,
		UpdatedAt: list.UpdatedAt,
	}, nil
}

func (s *ListStorage) GetListsByUserID(ctx context.Context, userID string) ([]model.List, error) {
	const op = "list.storage.GetListsByUserID"

	items, err := s.Queries.GetListsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get lists: %w", op, err)
	}
	if len(items) == 0 {
		return nil, le.ErrNoListsFound
	}

	var lists []model.List

	for _, item := range items {
		lists = append(lists, model.List{
			ID:        item.ID,
			Title:     item.Title,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return lists, nil
}

func (s *ListStorage) GetDefaultListID(ctx context.Context, userID string) (string, error) {
	const op = "list.storage.GetDefaultListID"

	listID, err := s.Queries.GetDefaultListID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", le.ErrNoListsFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: failed to get default list: %w", op, err)
	}
	return listID, nil
}

func (s *ListStorage) UpdateList(ctx context.Context, list model.List) error {
	const op = "list.storage.UpdateList"

	_, err := s.Queries.UpdateList(ctx, sqlc.UpdateListParams{
		Title:     list.Title,
		UpdatedAt: list.UpdatedAt,
		ID:        list.ID,
		UserID:    list.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrListNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to update list: %w", op, err)
	}
	return nil
}

func (s *ListStorage) DeleteList(ctx context.Context, list model.List) error {
	const op = "list.storage.DeleteList"

	err := s.Queries.DeleteList(ctx, sqlc.DeleteListParams{
		ID:     list.ID,
		UserID: list.UserID,
		DeletedAt: pgtype.Timestamptz{
			Time:  list.DeletedAt,
			Valid: true,
		},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrListNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to delete list: %w", op, err)
	}
	return nil
}
