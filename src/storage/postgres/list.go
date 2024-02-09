package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/models"
	"time"
)

type ListStorage struct {
	*pgxpool.Pool
}

func NewListStorage(pool *pgxpool.Pool) *ListStorage {
	return &ListStorage{Pool: pool}
}

func (s *ListStorage) CreateList(ctx context.Context, list models.List) error {
	const (
		op = "list.storage.CreateList"

		query = `
			INSERT INTO lists (id, title, user_id, updated_at)
			VALUES ($1, $2, $3, $4)`
	)

	_, err := s.Exec(
		ctx,
		query,
		list.ID,
		list.Title,
		list.UserID,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("%s: failed to insert new list: %w", op, err)
	}

	return nil
}

func (s *ListStorage) GetListByID(ctx context.Context, listID, userID string) (models.List, error) {
	const (
		op = "list.storage.GetListByID"

		query = `
			SELECT id, title, user_id, updated_at
			FROM lists
			WHERE id = $1
			  AND user_id = $2
			  AND deleted_at IS NULL`
	)

	var list models.List

	err := s.QueryRow(ctx, query, listID, userID).Scan(
		&list.Title,
		&list.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.List{}, c.ErrListNotFound
	}
	if err != nil {
		return models.List{}, fmt.Errorf("%s: failed to get list: %w", op, err)
	}

	list.ID = listID

	return list, nil

}

func (s *ListStorage) GetListsByUserID(ctx context.Context, userID string) ([]models.List, error) {
	const (
		op = "list.storage.GetListsByUserID"

		query = `
			SELECT id, title, updated_at
			FROM lists
			WHERE user_id = $1
			  AND deleted_at IS NULL
			ORDER BY id`
	)

	rows, err := s.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	defer rows.Close()

	var lists []models.List

	for rows.Next() {
		list := models.List{}

		err = rows.Scan(
			&list.ID,
			&list.Title,
			&list.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		lists = append(lists, list)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(lists) == 0 {
		return nil, c.ErrNoListsFound
	}

	return lists, nil
}

func (s *ListStorage) UpdateList(ctx context.Context, list models.List) error {
	const (
		op = "list.storage.UpdateList"

		query = `
			UPDATE lists
			SET title = $1,	updated_at = $2
			WHERE id = $3
			  AND user_id = $4`
	)

	result, err := s.Exec(
		ctx,
		query,
		list.Title,
		time.Now().UTC(),
		list.ID,
		list.UserID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to update list: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return c.ErrListNotFound
	}

	return nil
}

func (s *ListStorage) DeleteList(ctx context.Context, listID, userID string) error {
	const (
		op = "list.storage.DeleteList"

		query = `
			UPDATE lists
			SET deleted_at = $1
			WHERE id = $2
			  AND user_id = $3
			  AND deleted_at IS NULL`
	)

	result, err := s.Exec(
		ctx,
		query,
		time.Now().UTC(),
		listID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to delete list: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return c.ErrListNotFound
	}

	return nil
}
