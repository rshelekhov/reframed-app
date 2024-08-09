package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshelekhov/reframed/internal/port"
	"github.com/rshelekhov/reframed/internal/storage/postgres/sqlc"
)

type UserStorage struct {
	*pgxpool.Pool
	*sqlc.Queries
}

func NewUserStorage(pool *pgxpool.Pool) port.UserStorage {
	return &UserStorage{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}

func (s *UserStorage) DeleteUserData(ctx context.Context, userID string) error {
	const op = "storage.UserStorage.DeleteUserData"

	err := s.Queries.DeleteUserRelatedData(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: failed to delete user related data: %w", op, err)
	}

	return nil
}
