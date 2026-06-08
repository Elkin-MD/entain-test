package request

import (
	"entaintest/internal/common/enum"
	repomodel "entaintest/internal/repository/model"
)

// ProcessTransactionRequest is the service-level input for a transaction.
type ProcessTransactionRequest struct {
	UserID        uint64
	SourceType    string
	State         string
	Amount        string
	TransactionID string
}

// ToTransaction builds the repository transaction from validated values.
func (r ProcessTransactionRequest) ToTransaction(state enum.TransactionState, sourceType enum.SourceType, amount int64) repomodel.Transaction {
	return repomodel.Transaction{
		TransactionID: r.TransactionID,
		UserID:        r.UserID,
		State:         state,
		SourceType:    sourceType,
		Amount:        amount,
	}
}
