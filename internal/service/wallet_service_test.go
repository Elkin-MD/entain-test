package service_test

import (
	"context"
	"testing"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/common/enum"
	"entaintest/internal/model"
	"entaintest/internal/model/request"
	"entaintest/internal/service"
	"entaintest/internal/service/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type WalletServiceSuite struct {
	suite.Suite

	repo    *mocks.MockWalletRepo
	service *service.WalletService

	validRequest request.TransactionRequest
}

func Test_WalletServiceSuite(t *testing.T) {
	suite.Run(t, new(WalletServiceSuite))
}

func (s *WalletServiceSuite) SetupTest() {
	s.repo = mocks.NewMockWalletRepo(s.T())
	s.service = service.New(s.repo)

	s.validRequest = request.TransactionRequest{
		UserID:        1,
		SourceType:    "game",
		State:         "win",
		Amount:        "10.15",
		TransactionID: "tx-1",
	}
}

func (s *WalletServiceSuite) Test_ProcessTransaction_ReturnsBalance_WhenRequestIsValid() {
	// Arrange
	expected := model.Transaction{
		TransactionID: "tx-1",
		UserID:        1,
		State:         enum.StateWin,
		SourceType:    enum.SourceGame,
		Amount:        1015,
	}
	s.repo.EXPECT().
		ApplyTransaction(mock.Anything, expected).
		Return(int64(1015), nil).
		Once()

	// Act
	balance, err := s.service.ProcessTransaction(context.Background(), s.validRequest)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(int64(1015), balance)
}

func (s *WalletServiceSuite) Test_ProcessTransaction_ReturnsInvalidInput_WhenRequestIsInvalid() {
	tests := []struct {
		name   string
		mutate func(req *request.TransactionRequest)
	}{
		{name: "bad_state", mutate: func(req *request.TransactionRequest) { req.State = "draw" }},
		{name: "bad_source_type", mutate: func(req *request.TransactionRequest) { req.SourceType = "carrier" }},
		{name: "bad_amount", mutate: func(req *request.TransactionRequest) { req.Amount = "1.234" }},
		{name: "empty_transaction_id", mutate: func(req *request.TransactionRequest) { req.TransactionID = "" }},
		{name: "negative_amount", mutate: func(req *request.TransactionRequest) { req.Amount = "-1.00" }},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			req := s.validRequest
			tt.mutate(&req)

			// Act
			_, err := s.service.ProcessTransaction(context.Background(), req)

			// Assert
			s.Require().ErrorIs(err, businesserror.ErrInvalidInput)
		})
	}
}

func (s *WalletServiceSuite) Test_ProcessTransaction_ReturnsRepositoryError_WhenRepositoryFails() {
	// Arrange
	s.repo.EXPECT().
		ApplyTransaction(mock.Anything, mock.Anything).
		Return(int64(0), businesserror.ErrInsufficientFunds).
		Once()

	// Act
	_, err := s.service.ProcessTransaction(context.Background(), s.validRequest)

	// Assert
	s.Require().ErrorIs(err, businesserror.ErrInsufficientFunds)
}

func (s *WalletServiceSuite) Test_GetBalance_ReturnsBalance_WhenUserExists() {
	// Arrange
	s.repo.EXPECT().
		GetBalance(mock.Anything, uint64(1)).
		Return(int64(925), nil).
		Once()

	// Act
	balance, err := s.service.GetBalance(context.Background(), 1)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(int64(925), balance)
}
