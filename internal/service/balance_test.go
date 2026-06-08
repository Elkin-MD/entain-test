package service_test

import (
	"context"
	"testing"

	"entaintest/internal/common/businesserror"
	repomodel "entaintest/internal/repository/model"
	"entaintest/internal/service"
	"entaintest/internal/service/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BalanceServiceSuite struct {
	suite.Suite

	users   *mocks.MockUserRepository
	service *service.BalanceService
}

func Test_BalanceServiceSuite(t *testing.T) {
	suite.Run(t, new(BalanceServiceSuite))
}

func (s *BalanceServiceSuite) SetupTest() {
	s.users = mocks.NewMockUserRepository(s.T())
	s.service = service.NewBalanceService(s.users)
}

func (s *BalanceServiceSuite) Test_GetBalance_ReturnsBalance_InCaseUserExists() {
	// Arrange
	s.users.EXPECT().
		GetBalance(mock.Anything, uint64(1)).
		Return(&repomodel.User{ID: 1, Balance: 925}, nil).
		Once()

	// Act
	response, err := s.service.GetBalance(context.Background(), 1)

	// Assert
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), response.UserID)
	s.Require().Equal(int64(925), response.Balance)
}

func (s *BalanceServiceSuite) Test_GetBalance_ReturnsError_InCaseUserMissing() {
	// Arrange
	s.users.EXPECT().
		GetBalance(mock.Anything, uint64(9)).
		Return(nil, businesserror.ErrUserNotFound).
		Once()

	// Act
	_, err := s.service.GetBalance(context.Background(), 9)

	// Assert
	s.Require().ErrorIs(err, businesserror.ErrUserNotFound)
}
