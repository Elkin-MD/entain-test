package service_test

import (
	"context"
	"testing"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/common/enum"
	repomodel "entaintest/internal/repository/model"
	"entaintest/internal/service"
	"entaintest/internal/service/mocks"
	servicerequest "entaintest/internal/service/model/request"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TransactionServiceSuite struct {
	suite.Suite

	tx           *mocks.MockTx
	balances     *mocks.MockBalanceRepository
	transactions *mocks.MockTransactionRepository
	service      *service.TransactionService

	validRequest servicerequest.ProcessTransactionRequest
}

func Test_TransactionServiceSuite(t *testing.T) {
	suite.Run(t, new(TransactionServiceSuite))
}

func (s *TransactionServiceSuite) SetupTest() {
	s.tx = mocks.NewMockTx(s.T())
	s.balances = mocks.NewMockBalanceRepository(s.T())
	s.transactions = mocks.NewMockTransactionRepository(s.T())
	s.service = service.NewTransactionService(s.tx, s.balances, s.transactions)

	s.validRequest = servicerequest.ProcessTransactionRequest{
		UserID:        1,
		SourceType:    "game",
		State:         "win",
		Amount:        "10.15",
		TransactionID: "tx-1",
	}
}

// expectTransaction makes the tx mock execute the passed function.
func (s *TransactionServiceSuite) expectTransaction() {
	s.tx.EXPECT().
		WithTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).
		Once()
}

func (s *TransactionServiceSuite) Test_ProcessTransaction_ReturnsResponse_InCaseWin() {
	// Arrange
	expectedTransaction := repomodel.Transaction{
		TransactionID: "tx-1",
		UserID:        1,
		State:         enum.StateWin,
		SourceType:    enum.SourceGame,
		Amount:        1015,
	}

	s.expectTransaction()
	s.balances.EXPECT().LockBalance(mock.Anything, uint64(1)).Return(int64(500), nil).Once()
	s.transactions.EXPECT().Insert(mock.Anything, expectedTransaction).Return(nil).Once()
	s.balances.EXPECT().UpdateBalance(mock.Anything, uint64(1), int64(1515)).Return(nil).Once()

	// Act
	response, err := s.service.ProcessTransaction(context.Background(), s.validRequest)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(int64(1015), response.Amount)
	s.Require().Equal(int64(1515), response.Balance)
	s.Require().Equal(enum.StateWin, response.State)
}

func (s *TransactionServiceSuite) Test_ProcessTransaction_ReturnsInsufficientFunds_InCaseBalanceTooLow() {
	// Arrange
	request := s.validRequest
	request.State = "lose"
	request.Amount = "10.00"

	expectedTransaction := repomodel.Transaction{
		TransactionID: "tx-1",
		UserID:        1,
		State:         enum.StateLose,
		SourceType:    enum.SourceGame,
		Amount:        1000,
	}

	s.expectTransaction()
	s.balances.EXPECT().LockBalance(mock.Anything, uint64(1)).Return(int64(500), nil).Once()
	s.transactions.EXPECT().Insert(mock.Anything, expectedTransaction).Return(nil).Once()

	// Act
	_, err := s.service.ProcessTransaction(context.Background(), request)

	// Assert
	s.Require().ErrorIs(err, businesserror.ErrInsufficientFunds)
}

func (s *TransactionServiceSuite) Test_ProcessTransaction_ReturnsError_InCaseUserMissing() {
	// Arrange
	s.expectTransaction()
	s.balances.EXPECT().LockBalance(mock.Anything, uint64(1)).Return(int64(0), businesserror.ErrUserNotFound).Once()

	// Act
	_, err := s.service.ProcessTransaction(context.Background(), s.validRequest)

	// Assert
	s.Require().ErrorIs(err, businesserror.ErrUserNotFound)
}

func (s *TransactionServiceSuite) Test_ProcessTransaction_ReturnsValidationError_InCaseRequestIsInvalid() {
	tests := []struct {
		name    string
		mutate  func(request *servicerequest.ProcessTransactionRequest)
		wantErr error
	}{
		{
			name:    "bad_state",
			mutate:  func(request *servicerequest.ProcessTransactionRequest) { request.State = "draw" },
			wantErr: businesserror.ErrInvalidState,
		},
		{
			name:    "bad_source_type",
			mutate:  func(request *servicerequest.ProcessTransactionRequest) { request.SourceType = "carrier" },
			wantErr: businesserror.ErrInvalidSourceType,
		},
		{
			name:    "empty_transaction_id",
			mutate:  func(request *servicerequest.ProcessTransactionRequest) { request.TransactionID = "" },
			wantErr: businesserror.ErrMissingTransactionID,
		},
		{
			name:    "bad_amount",
			mutate:  func(request *servicerequest.ProcessTransactionRequest) { request.Amount = "1.234" },
			wantErr: businesserror.ErrInvalidAmount,
		},
		{
			name:    "negative_amount",
			mutate:  func(request *servicerequest.ProcessTransactionRequest) { request.Amount = "-1.00" },
			wantErr: businesserror.ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			request := s.validRequest
			tt.mutate(&request)

			// Act
			_, err := s.service.ProcessTransaction(context.Background(), request)

			// Assert
			s.Require().ErrorIs(err, tt.wantErr)
		})
	}
}
