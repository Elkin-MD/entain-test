package enum

import (
	"database/sql/driver"
	"fmt"
)

// SourceType identifies the origin of a transaction.
type SourceType string

const (
	SourceGame    SourceType = "game"
	SourceServer  SourceType = "server"
	SourcePayment SourceType = "payment"
)

var (
	sourceTypeID = map[SourceType]int64{
		SourceGame:    1,
		SourceServer:  2,
		SourcePayment: 3,
	}
	sourceTypeName = map[int64]SourceType{
		1: SourceGame,
		2: SourceServer,
		3: SourcePayment,
	}
)

// Valid reports whether the source type is one of the supported values.
func (s SourceType) Valid() bool {
	_, ok := sourceTypeID[s]

	return ok
}

// Value converts the source type to its numeric id for storage.
func (s SourceType) Value() (driver.Value, error) {
	id, ok := sourceTypeID[s]
	if !ok {
		return nil, fmt.Errorf("invalid source type %q", s)
	}

	return id, nil
}

// Scan converts a numeric id from storage back to the source type.
func (s *SourceType) Scan(src any) error {
	id, ok := src.(int64)
	if !ok {
		return fmt.Errorf("cannot scan %T into SourceType", src)
	}

	sourceType, ok := sourceTypeName[id]
	if !ok {
		return fmt.Errorf("unknown source type id %d", id)
	}

	*s = sourceType

	return nil
}
