package service

import (
	"context"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/common/enum"
	"entaintest/internal/money"
	repomodel "entaintest/internal/repository/model"
	servicerequest "entaintest/internal/service/model/request"
	serviceresponse "entaintest/internal/service/model/response"
)

// Tx runs a function within a single database transaction.
type Tx interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// BalanceRepository reads and writes a user's balance under a row lock.
type BalanceRepository interface {
	LockBalance(ctx context.Context, userID uint64) (int64, error)
	UpdateBalance(ctx context.Context, userID uint64, balance int64) error
}

// TransactionRepository records transactions.
type TransactionRepository interface {
	Insert(ctx context.Context, transaction repomodel.Transaction) error
}

// TransactionService applies transactions and recalculates balances.
type TransactionService struct {
	tx           Tx
	balances     BalanceRepository
	transactions TransactionRepository
}

// NewTransactionService creates a TransactionService.
func NewTransactionService(
	tx Tx,
	balances BalanceRepository,
	transactions TransactionRepository,
) *TransactionService {
	return &TransactionService{
		tx:           tx,
		balances:     balances,
		transactions: transactions,
	}
}

// ProcessTransaction validates the request and applies it atomically, returning the resulting transaction and balance.
func (s *TransactionService) ProcessTransaction(ctx context.Context, request servicerequest.ProcessTransactionRequest) (*serviceresponse.ProcessTransactionResponse, error) {
	state, sourceType, amount, err := validate(request)
	if err != nil {
		return nil, err
	}

	transaction := request.ToTransaction(state, sourceType, amount)

	var balance int64

	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		current, err := s.balances.LockBalance(ctx, transaction.UserID)
		if err != nil {
			return err
		}

		if err := s.transactions.Insert(ctx, transaction); err != nil {
			return err
		}

		balance, err = nextBalance(current, transaction)
		if err != nil {
			return err
		}

		return s.balances.UpdateBalance(ctx, transaction.UserID, balance)
	})
	if err != nil {
		return nil, err
	}

	return &serviceresponse.ProcessTransactionResponse{
		TransactionID: transaction.TransactionID,
		UserID:        transaction.UserID,
		State:         state,
		SourceType:    sourceType,
		Amount:        amount,
		Balance:       balance,
	}, nil
}

func validate(request servicerequest.ProcessTransactionRequest) (enum.TransactionState, enum.SourceType, int64, error) {
	state := enum.TransactionState(request.State)
	if !state.Valid() {
		return "", "", 0, businesserror.ErrInvalidState
	}

	sourceType := enum.SourceType(request.SourceType)
	if !sourceType.Valid() {
		return "", "", 0, businesserror.ErrInvalidSourceType
	}

	if request.TransactionID == "" {
		return "", "", 0, businesserror.ErrMissingTransactionID
	}

	amount, err := money.Parse(request.Amount)
	if err != nil {
		return "", "", 0, businesserror.ErrInvalidAmount
	}

	return state, sourceType, amount, nil
}

func nextBalance(balance int64, transaction repomodel.Transaction) (int64, error) {
	if transaction.State == enum.StateWin {
		return balance + transaction.Amount, nil
	}

	if transaction.Amount > balance {
		return 0, businesserror.ErrInsufficientFunds
	}

	return balance - transaction.Amount, nil
}
