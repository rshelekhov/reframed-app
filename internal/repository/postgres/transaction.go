package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshelekhov/reframed/internal/port"
)

type Executor struct {
	*pgxpool.Pool
}

func NewExecutor(pool *pgxpool.Pool) port.StorageExecutor {
	return &Executor{Pool: pool}
}

type txKey struct{}

// injectTx injects transaction into context
func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// extractTx extracts transaction from context
func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

func (e *Executor) ExecSQL(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	tx := extractTx(ctx)
	if tx != nil {
		return tx.Exec(ctx, sql, arguments...)
	}
	return e.Pool.Exec(ctx, sql, arguments...)
}

// Add methods for QueryRow, Query

type TransactionManager struct {
	*pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{Pool: pool}
}

func (tm *TransactionManager) WithinTransaction(ctx context.Context, op string, tFunc func(ctx context.Context) error) error {
	// Begin transaction
	tx, err := tm.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = fmt.Errorf("%s: failed to rollback transaction: %w", op, errRollback)
			}
		} else {
			if errCommit := tx.Commit(ctx); errCommit != nil {
				err = fmt.Errorf("%s: failed to commit transaction: %w", op, errCommit)
			}
		}
	}()

	err = tFunc(injectTx(ctx, tx))
	return err

}
