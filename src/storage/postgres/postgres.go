// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rshelekhov/reframed/config"
	"net"
)

type Storage struct {
	DB *sqlx.DB
}

// NewStorage creates a new Postgres storage
func NewStorage(cfg *config.Config) (*pgxpool.Pool, error) {
	const op = "storage.postgres.NewStorage"

	poolCfg, err := pgxpool.ParseConfig(cfg.Postgres.ConnURL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse config: %w", op, err)
	}

	poolCfg.MaxConnLifetime = cfg.Postgres.IdleTimeout
	poolCfg.MaxConns = int32(cfg.Postgres.ConnPoolSize)

	dialer := &net.Dialer{KeepAlive: cfg.Postgres.DialTimeout}
	dialer.Timeout = cfg.Postgres.DialTimeout
	poolCfg.ConnConfig.DialFunc = dialer.DialContext

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create pgx connection pool: %w", op, err)
	}

	return pool, nil
}

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
