package businesserror

// Error is a comparable, constant business error.
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrUserNotFound         Error = "user not found"
	ErrInsufficientFunds    Error = "insufficient funds"
	ErrDuplicateTransaction Error = "transaction already processed"
	ErrInvalidState         Error = "invalid state"
	ErrInvalidSourceType    Error = "invalid source type"
	ErrInvalidAmount        Error = "invalid amount"
	ErrMissingTransactionID Error = "missing transactionId"
)
