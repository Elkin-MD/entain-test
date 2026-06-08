package enum

import (
	"database/sql/driver"
	"fmt"
)

// TransactionState is the effect a transaction has on a balance.
type TransactionState string

const (
	StateWin  TransactionState = "win"
	StateLose TransactionState = "lose"
)

var (
	transactionStateID = map[TransactionState]int64{
		StateWin:  1,
		StateLose: 2,
	}
	transactionStateName = map[int64]TransactionState{
		1: StateWin,
		2: StateLose,
	}
)

// Valid reports whether the state is one of the supported values.
func (s TransactionState) Valid() bool {
	_, ok := transactionStateID[s]

	return ok
}

// Value converts the state to its numeric id for storage.
func (s TransactionState) Value() (driver.Value, error) {
	id, ok := transactionStateID[s]
	if !ok {
		return nil, fmt.Errorf("invalid transaction state %q", s)
	}

	return id, nil
}

// Scan converts a numeric id from storage back to the state.
func (s *TransactionState) Scan(src any) error {
	id, ok := src.(int64)
	if !ok {
		return fmt.Errorf("cannot scan %T into TransactionState", src)
	}

	state, ok := transactionStateName[id]
	if !ok {
		return fmt.Errorf("unknown transaction state id %d", id)
	}

	*s = state

	return nil
}
