package model

// User is an account with a balance stored in minor units (cents).
type User struct {
	ID      uint64
	Balance int64
}
