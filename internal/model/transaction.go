package model

import "entaintest/internal/common/enum"

// Transaction is a single balance-changing operation, with Amount in cents.
type Transaction struct {
	TransactionID string
	UserID        uint64
	State         enum.State
	SourceType    enum.SourceType
	Amount        int64
}
