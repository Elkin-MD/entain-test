package repository

import (
	"context"
	"errors"
	"fmt"

	"entaintest/internal/common/businesserror"
	model "entaintest/internal/repository/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository accesses the users table.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a UserRepository backed by the given pool.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

// GetBalance returns the user, or ErrUserNotFound.
func (r *UserRepository) GetBalance(ctx context.Context, userID uint64) (*model.User, error) {
	user := model.User{
		ID: userID,
	}

	err := querierFrom(ctx, r.pool).QueryRow(
		ctx,
		`SELECT balance
		 FROM users
		 WHERE id = @id`,
		pgx.NamedArgs{"id": userID},
	).Scan(&user.Balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, businesserror.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	return &user, nil
}

// LockBalance reads and row-locks the balance, giving a consistent lock order
// that avoids a lock-upgrade deadlock under concurrency.
func (r *UserRepository) LockBalance(ctx context.Context, userID uint64) (int64, error) {
	var balance int64

	err := querierFrom(ctx, r.pool).QueryRow(
		ctx,
		`SELECT balance
		 FROM users
		 WHERE id = @id
		 FOR UPDATE`,
		pgx.NamedArgs{"id": userID},
	).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, businesserror.ErrUserNotFound
	}

	if err != nil {
		return 0, fmt.Errorf("lock user: %w", err)
	}

	return balance, nil
}

// UpdateBalance sets the user's balance.
func (r *UserRepository) UpdateBalance(ctx context.Context, userID uint64, balance int64) error {
	_, err := querierFrom(ctx, r.pool).Exec(
		ctx,
		`UPDATE users
		 SET balance = @balance
		 WHERE id = @id`,
		pgx.NamedArgs{
			"balance": balance,
			"id":      userID,
		},
	)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	return nil
}
