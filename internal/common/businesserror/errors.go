package businesserror

// Error is a comparable, constant business error.
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrUserNotFound         Error = "user not found"
	ErrInsufficientFunds    Error = "insufficient funds"
	ErrDuplicateTransaction Error = "duplicate transaction"
	ErrInvalidInput         Error = "invalid input"
)
