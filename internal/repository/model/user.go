package model

// User is an account with a balance in minor units (cents).
type User struct {
	ID      uint64
	Balance int64
}
