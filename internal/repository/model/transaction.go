package model

import "entaintest/internal/common/enum"

// Transaction is a balance-changing operation as stored by the repository.
type Transaction struct {
	TransactionID string
	UserID        uint64
	State         enum.TransactionState
	SourceType    enum.SourceType
	Amount        int64
}
