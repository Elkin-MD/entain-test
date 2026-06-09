package service

import (
	"context"

	repomodel "entaintest/internal/repository/model"
	serviceresponse "entaintest/internal/service/model/response"
)

// UserRepository reads a user's balance.
type UserRepository interface {
	GetBalance(ctx context.Context, userID uint64) (*repomodel.User, error)
}

// BalanceService reads user balances.
type BalanceService struct {
	users UserRepository
}

// NewBalanceService creates a BalanceService.
func NewBalanceService(users UserRepository) *BalanceService {
	return &BalanceService{
		users: users,
	}
}

// GetBalance returns the user's current balance.
func (s *BalanceService) GetBalance(ctx context.Context, userID uint64) (*serviceresponse.GetBalanceResponse, error) {
	user, err := s.users.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &serviceresponse.GetBalanceResponse{
		UserID:  user.ID,
		Balance: user.Balance,
	}, nil
}
