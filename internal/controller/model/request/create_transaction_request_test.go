package request

import (
	"testing"

	"entaintest/internal/common/businesserror"

	"github.com/stretchr/testify/require"
)

func Test_NewCreateTransactionRequest_ReturnsRequest_InCaseFieldsAreValid(t *testing.T) {
	// Act
	request, err := NewCreateTransactionRequest(1, "game", "win", "10.15", "tx-1")

	// Assert
	require.NoError(t, err)
	require.Equal(t, uint64(1), request.UserID)
	require.Equal(t, "game", request.SourceType)
	require.Equal(t, "win", request.State)
	require.Equal(t, "10.15", request.Amount)
	require.Equal(t, "tx-1", request.TransactionID)
}

func Test_NewCreateTransactionRequest_ReturnsError_InCaseFieldsAreInvalid(t *testing.T) {
	tests := []struct {
		name          string
		sourceType    string
		state         string
		amount        string
		transactionID string
		wantErr       error
	}{
		{name: "bad_state", sourceType: "game", state: "draw", amount: "1.00", transactionID: "t", wantErr: businesserror.ErrInvalidState},
		{name: "bad_source_type", sourceType: "carrier", state: "win", amount: "1.00", transactionID: "t", wantErr: businesserror.ErrInvalidSourceType},
		{name: "empty_transaction_id", sourceType: "game", state: "win", amount: "1.00", transactionID: "", wantErr: businesserror.ErrMissingTransactionID},
		{name: "bad_amount", sourceType: "game", state: "win", amount: "1.234", transactionID: "t", wantErr: businesserror.ErrInvalidAmount},
		{name: "negative_amount", sourceType: "game", state: "win", amount: "-1.00", transactionID: "t", wantErr: businesserror.ErrInvalidAmount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := NewCreateTransactionRequest(1, tt.sourceType, tt.state, tt.amount, tt.transactionID)

			// Assert
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
