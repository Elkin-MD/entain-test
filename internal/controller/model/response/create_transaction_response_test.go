package response

import (
	"testing"

	"entaintest/internal/common/enum"
	serviceresponse "entaintest/internal/service/model/response"

	"github.com/stretchr/testify/require"
)

func Test_NewCreateTransactionResponse_ReturnsResponse_InCaseGivenServiceResponse(t *testing.T) {
	// Arrange
	source := &serviceresponse.ProcessTransactionResponse{
		TransactionID: "tx-1",
		UserID:        1,
		State:         enum.StateWin,
		SourceType:    enum.SourceGame,
		Amount:        1015,
		Balance:       2030,
	}

	// Act
	response := NewCreateTransactionResponse(source)

	// Assert
	require.Equal(t, "tx-1", response.TransactionID)
	require.Equal(t, uint64(1), response.UserID)
	require.Equal(t, "win", response.State)
	require.Equal(t, "game", response.SourceType)
	require.Equal(t, "10.15", response.Amount)
	require.Equal(t, "20.30", response.Balance)
}
