package service

import (
	"context"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/common/enum"
	"entaintest/internal/model"
	"entaintest/internal/model/request"
	"entaintest/internal/money"
)

// WalletRepo is the persistence behaviour the service depends on.
type WalletRepo interface {
	ApplyTransaction(ctx context.Context, transaction model.Transaction) (int64, error)
	GetBalance(ctx context.Context, userID uint64) (int64, error)
}

// WalletService applies the wallet business rules.
type WalletService struct {
	repo WalletRepo
}

// New creates a WalletService over the given repository.
func New(repo WalletRepo) *WalletService {
	return &WalletService{repo: repo}
}

// ProcessTransaction validates the request and applies it, returning the new balance in cents.
func (s *WalletService) ProcessTransaction(ctx context.Context, request request.TransactionRequest) (int64, error) {
	state := enum.State(request.State)
	if !state.Valid() {
		return 0, businesserror.ErrInvalidInput
	}

	sourceType := enum.SourceType(request.SourceType)
	if !sourceType.Valid() {
		return 0, businesserror.ErrInvalidInput
	}

	if request.TransactionID == "" {
		return 0, businesserror.ErrInvalidInput
	}

	amount, err := money.Parse(request.Amount)
	if err != nil {
		return 0, businesserror.ErrInvalidInput
	}

	return s.repo.ApplyTransaction(ctx, model.Transaction{
		TransactionID: request.TransactionID,
		UserID:        request.UserID,
		State:         state,
		SourceType:    sourceType,
		Amount:        amount,
	})
}

// GetBalance returns the user's current balance in cents.
func (s *WalletService) GetBalance(ctx context.Context, userID uint64) (int64, error) {
	return s.repo.GetBalance(ctx, userID)
}
