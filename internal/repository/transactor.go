package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

// querier is the subset of *pgxpool.Pool and pgx.Tx the repositories use.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Transactor runs a function within a single database transaction, sharing the
// transaction with repositories through the context.
type Transactor struct {
	pool *pgxpool.Pool
}

// NewTransactor creates a Transactor backed by the given pool.
func NewTransactor(pool *pgxpool.Pool) *Transactor {
	return &Transactor{
		pool: pool,
	}
}

// WithTx runs fn inside a transaction, committing on success and rolling back on error.
func (t *Transactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback(ctx)

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// querierFrom returns the transaction carried by ctx, or the pool.
func querierFrom(ctx context.Context, pool *pgxpool.Pool) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return pool
}
