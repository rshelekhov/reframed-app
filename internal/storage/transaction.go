package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BeginTransaction begins a new transaction
func BeginTransaction(pool *pgxpool.Pool, ctx context.Context, op string) (pgx.Tx, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	return tx, nil
}

// RollbackOnError rolls back the transaction if an error occurred
func RollbackOnError(err *error, tx pgx.Tx, ctx context.Context, op string) {
	if errRollback := tx.Rollback(ctx); errRollback != nil {
		*err = fmt.Errorf("%s: failed to rollback transaction: %w", op, errRollback)
	}
}

// CommitTransaction commits the transaction
func CommitTransaction(err *error, tx pgx.Tx, ctx context.Context, op string) {
	if errCommit := tx.Commit(ctx); err != nil {
		*err = fmt.Errorf("%s: failed to commit transaction: %w", op, errCommit)
	}
}
