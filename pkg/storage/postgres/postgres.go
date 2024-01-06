// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"fmt"
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

	/*
		db, err := sqlx.Open("pgx", dsn)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if err = db.Ping(); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return &Storage{db}, nil*/
}
