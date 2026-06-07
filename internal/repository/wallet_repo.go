package repository

import (
	"context"
	"errors"
	"fmt"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/common/enum"
	"entaintest/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

// WalletRepository persists users and transactions in Postgres.
type WalletRepository struct {
	pool *pgxpool.Pool
}

// NewWalletRepository creates a repository backed by the given pool.
func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

// GetBalance returns the current balance in cents for a user.
func (r *WalletRepository) GetBalance(ctx context.Context, userID uint64) (int64, error) {
	var balance int64

	err := r.pool.QueryRow(ctx,
		`SELECT balance
		 FROM users
		 WHERE id = @id`,
		pgx.NamedArgs{"id": userID},
	).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, businesserror.ErrUserNotFound
	}

	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}

	return balance, nil
}

// ApplyTransaction atomically records a transaction and updates the user balance, returning the new balance in cents.
func (r *WalletRepository) ApplyTransaction(ctx context.Context, transaction model.Transaction) (int64, error) {
	dbTx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}

	defer dbTx.Rollback(ctx)

	// Lock the user row first to serialize concurrent transactions and avoid a lock-upgrade deadlock.
	var balance int64

	err = dbTx.QueryRow(ctx,
		`SELECT balance
		 FROM users
		 WHERE id = @id
		 FOR UPDATE`,
		pgx.NamedArgs{"id": transaction.UserID},
	).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, businesserror.ErrUserNotFound
	}

	if err != nil {
		return 0, fmt.Errorf("lock user: %w", err)
	}

	err = r.insertTransaction(ctx, dbTx, transaction)
	if err != nil {
		return 0, err
	}

	newBalance := balance

	switch transaction.State {
	case enum.StateWin:
		newBalance += transaction.Amount
	case enum.StateLose:
		if transaction.Amount > balance {
			return 0, businesserror.ErrInsufficientFunds
		}

		newBalance -= transaction.Amount
	default:
		return 0, fmt.Errorf("unsupported state: %q", transaction.State)
	}

	_, err = dbTx.Exec(ctx,
		`UPDATE users
		 SET balance = @balance
		 WHERE id = @id`,
		pgx.NamedArgs{
			"balance": newBalance,
			"id":      transaction.UserID,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("update balance: %w", err)
	}

	if err = dbTx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return newBalance, nil
}

func (r *WalletRepository) insertTransaction(ctx context.Context, dbTx pgx.Tx, transaction model.Transaction) error {
	_, err := dbTx.Exec(ctx,
		`INSERT INTO transactions (transaction_id, user_id, state, source_type, amount)
		 VALUES (@transaction_id, @user_id, @state, @source_type, @amount)`,
		pgx.NamedArgs{
			"transaction_id": transaction.TransactionID,
			"user_id":        transaction.UserID,
			"state":          string(transaction.State),
			"source_type":    string(transaction.SourceType),
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
