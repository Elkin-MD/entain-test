package controller_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"entaintest/internal/common/businesserror"
	"entaintest/internal/controller"
	"entaintest/internal/controller/mocks"
	serviceresponse "entaintest/internal/service/model/response"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BalanceControllerSuite struct {
	suite.Suite

	transactionService *mocks.MockTransactionService
	balanceService     *mocks.MockBalanceService
	controller         *controller.Controller
}

func Test_BalanceControllerSuite(t *testing.T) {
	suite.Run(t, new(BalanceControllerSuite))
}

func (s *BalanceControllerSuite) SetupTest() {
	s.transactionService = mocks.NewMockTransactionService(s.T())
	s.balanceService = mocks.NewMockBalanceService(s.T())
	s.controller = controller.New(s.transactionService, s.balanceService)
}

func (s *BalanceControllerSuite) newRequest(userID string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/user/"+userID+"/balance", nil)
	request.SetPathValue("userId", userID)

	return request
}

func (s *BalanceControllerSuite) Test_GetBalance_ReturnsBalance_InCaseUserExists() {
	// Arrange
	s.balanceService.EXPECT().
		GetBalance(mock.Anything, uint64(1)).
		Return(&serviceresponse.GetBalanceResponse{UserID: 1, Balance: 925}, nil).
		Once()

	// Act
	response, err := s.controller.GetBalance(httptest.NewRecorder(), s.newRequest("1"))

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), response.UserID)
	s.Require().Equal("9.25", response.Balance)
}

func (s *BalanceControllerSuite) Test_GetBalance_ReturnsInvalidUserID_InCaseUserIDIsNotPositive() {
	// Act
	_, err := s.controller.GetBalance(httptest.NewRecorder(), s.newRequest("abc"))

	// Assert
	s.Require().ErrorIs(err, controller.ErrInvalidUserID)
}

func (s *BalanceControllerSuite) Test_GetBalance_ReturnsServiceError_InCaseServiceFails() {
	// Arrange
	s.balanceService.EXPECT().
		GetBalance(mock.Anything, uint64(9)).
		Return(nil, businesserror.ErrUserNotFound).
		Once()

	// Act
	_, err := s.controller.GetBalance(httptest.NewRecorder(), s.newRequest("9"))

	// Assert
	s.Require().ErrorIs(err, businesserror.ErrUserNotFound)
}
