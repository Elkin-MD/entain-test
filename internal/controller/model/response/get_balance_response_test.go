package response

import (
	"testing"

	serviceresponse "entaintest/internal/service/model/response"

	"github.com/stretchr/testify/require"
)

func Test_NewGetBalanceResponse_ReturnsResponse_InCaseGivenServiceResponse(t *testing.T) {
	// Arrange
	source := &serviceresponse.GetBalanceResponse{
		UserID:  2,
		Balance: 925,
	}

	// Act
	response := NewGetBalanceResponse(source)

	// Assert
	require.Equal(t, uint64(2), response.UserID)
	require.Equal(t, "9.25", response.Balance)
}
