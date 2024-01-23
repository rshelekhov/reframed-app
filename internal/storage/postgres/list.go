package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshelekhov/reframed/internal/models"
)

type ListStorage struct {
	*pgxpool.Pool
}

func NewListStorage(pool *pgxpool.Pool) *ListStorage {
	return &ListStorage{Pool: pool}
}

func (l ListStorage) CreateList(ctx context.Context, list models.List) error {
	const (
		op = "list.storage.CreateList"

		query = `INSERT INTO lists (id, title, user_id, updated_at) VALUES ($1, $2, $3, $4)`
	)

	_, err := l.Exec(ctx, query, list.ID, list.Title, list.UserID, list.UpdatedAt)
	if err != nil {
		return fmt.Errorf("%s: failed to insert new list: %w", op, err)
	}

	return nil
}

func (l ListStorage) GetListByID(ctx context.Context, id string) (models.List, error) {
	//TODO implement me
	panic("implement me")
}

func (l ListStorage) GetLists(ctx context.Context, pgn models.Pagination) ([]models.List, error) {
	//TODO implement me
	panic("implement me")
}

func (l ListStorage) UpdateList(ctx context.Context, list models.List) error {
	//TODO implement me
	panic("implement me")
}

func (l ListStorage) DeleteList(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
