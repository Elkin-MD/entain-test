package controller_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/common/enum"
	"entaintest/internal/controller"
	"entaintest/internal/controller/mocks"
	servicerequest "entaintest/internal/service/model/request"
	serviceresponse "entaintest/internal/service/model/response"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TransactionControllerSuite struct {
	suite.Suite

	transactionService *mocks.MockTransactionService
	balanceService     *mocks.MockBalanceService
	controller         *controller.Controller
}

func Test_TransactionControllerSuite(t *testing.T) {
	suite.Run(t, new(TransactionControllerSuite))
}

func (s *TransactionControllerSuite) SetupTest() {
	s.transactionService = mocks.NewMockTransactionService(s.T())
	s.balanceService = mocks.NewMockBalanceService(s.T())
	s.controller = controller.New(s.transactionService, s.balanceService)
}

func (s *TransactionControllerSuite) newRequest(userID, body string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/user/"+userID+"/transaction", strings.NewReader(body))
	request.SetPathValue("userId", userID)
	request.Header.Set("Source-Type", "game")

	return request
}

func (s *TransactionControllerSuite) Test_CreateTransaction_ReturnsResponse_InCaseRequestIsValid() {
	// Arrange
	expectedRequest := servicerequest.ProcessTransactionRequest{
		UserID:        1,
		SourceType:    "game",
		State:         "win",
		Amount:        "10.15",
		TransactionID: "tx-1",
	}

	s.transactionService.EXPECT().
		ProcessTransaction(mock.Anything, expectedRequest).
		Return(&serviceresponse.ProcessTransactionResponse{
			TransactionID: "tx-1",
			UserID:        1,
			State:         enum.StateWin,
			SourceType:    enum.SourceGame,
			Amount:        1015,
			Balance:       1015,
		}, nil).
		Once()

	// Act
	response, err := s.controller.CreateTransaction(
		httptest.NewRecorder(),
		s.newRequest("1", `{"state":"win","amount":"10.15","transactionId":"tx-1"}`),
	)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), response.UserID)
	s.Require().Equal("win", response.State)
	s.Require().Equal("10.15", response.Amount)
	s.Require().Equal("10.15", response.Balance)
}

func (s *TransactionControllerSuite) Test_CreateTransaction_ReturnsInvalidUserID_InCaseUserIDIsNotPositive() {
	// Act
	_, err := s.controller.CreateTransaction(
		httptest.NewRecorder(),
		s.newRequest("0", `{"state":"win","amount":"10.15","transactionId":"tx-1"}`),
	)

	// Assert
	s.Require().ErrorIs(err, controller.ErrInvalidUserID)
}

func (s *TransactionControllerSuite) Test_CreateTransaction_ReturnsInvalidBody_InCaseBodyIsMalformed() {
	// Act
	_, err := s.controller.CreateTransaction(httptest.NewRecorder(), s.newRequest("1", "not-json"))

	// Assert
	s.Require().ErrorIs(err, controller.ErrInvalidBody)
}

func (s *TransactionControllerSuite) Test_CreateTransaction_ReturnsServiceError_InCaseServiceFails() {
	// Arrange
	s.transactionService.EXPECT().
		ProcessTransaction(mock.Anything, mock.Anything).
		Return(nil, businesserror.ErrDuplicateTransaction).
		Once()

	// Act
	_, err := s.controller.CreateTransaction(
		httptest.NewRecorder(),
		s.newRequest("1", `{"state":"win","amount":"10.15","transactionId":"tx-1"}`),
	)

	// Assert
	s.Require().ErrorIs(err, businesserror.ErrDuplicateTransaction)
}
