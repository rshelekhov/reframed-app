package port

import (
	"context"
	"github.com/jackc/pgx/v5/pgconn"
)

type TransactionManager interface {
	WithinTransaction(ctx context.Context, op string, tFunc func(ctx context.Context) error) error
}

type StorageExecutor interface {
	ExecSQL(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}
