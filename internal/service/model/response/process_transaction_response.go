package response

import "entaintest/internal/common/enum"

// ProcessTransactionResponse is the service-level result of a transaction.
type ProcessTransactionResponse struct {
	TransactionID string
	UserID        uint64
	State         enum.TransactionState
	SourceType    enum.SourceType
	Amount        int64
	Balance       int64
}
