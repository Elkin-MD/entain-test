package repository

import (
	"context"
	"errors"
	"fmt"

	"entaintest/internal/common/businesserror"
	model "entaintest/internal/repository/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

// TransactionRepository accesses the transactions table.
type TransactionRepository struct {
	pool *pgxpool.Pool
}

// NewTransactionRepository creates a TransactionRepository backed by the given pool.
func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{
		pool: pool,
	}
}

// Insert records a transaction, returning ErrDuplicateTransaction for a repeat
// transactionId or ErrUserNotFound for an unknown user.
func (r *TransactionRepository) Insert(ctx context.Context, transaction model.Transaction) error {
	_, err := querierFrom(ctx, r.pool).Exec(
		ctx,
		`INSERT INTO transactions (transaction_id, user_id, state, source_type, amount)
		 VALUES (@transaction_id, @user_id, @state, @source_type, @amount)`,
		pgx.NamedArgs{
			"transaction_id": transaction.TransactionID,
			"user_id":        transaction.UserID,
			"state":          transaction.State,
			"source_type":    transaction.SourceType,
			"amount":         transaction.Amount,
		},
	)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case uniqueViolation:
			return businesserror.ErrDuplicateTransaction
		case foreignKeyViolation:
			return businesserror.ErrUserNotFound
		}
	}

	return fmt.Errorf("insert transaction: %w", err)
}
