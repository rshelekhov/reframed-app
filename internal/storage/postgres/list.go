package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
	"github.com/rshelekhov/reframed/pkg/constants/le"
)

type ListStorage struct {
	*pgxpool.Pool
	*Queries
}

func NewListStorage(pool *pgxpool.Pool) port.ListStorage {
	return &ListStorage{
		Pool:    pool,
		Queries: New(pool),
	}
}

func (s *ListStorage) CreateList(ctx context.Context, list model.List) error {
	const op = "list.storage.CreateList"

	if err := s.Queries.CreateList(ctx, CreateListParams{
		ID:        list.ID,
		Title:     list.Title,
		UserID:    list.UserID,
		UpdatedAt: list.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("%s: failed to create new list: %w", op, err)
	}

	return nil
}

func (s *ListStorage) GetListByID(ctx context.Context, listID, userID string) (model.List, error) {
	const op = "list.storage.GetListByID"

	list, err := s.Queries.GetListByID(ctx, GetListByIDParams{
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

func (s *ListStorage) UpdateList(ctx context.Context, list model.List) error {
	const op = "list.storage.UpdateList"

	err := s.Queries.UpdateList(ctx, UpdateListParams{
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

	err := s.Queries.DeleteList(ctx, DeleteListParams{
		ID:     list.ID,
		UserID: list.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrListNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to delete list: %w", op, err)
	}

	return nil
}
