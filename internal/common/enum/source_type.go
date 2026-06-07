package enum

// SourceType identifies the origin of a transaction.
type SourceType string

const (
	SourceGame    SourceType = "game"
	SourceServer  SourceType = "server"
	SourcePayment SourceType = "payment"
)

// Valid reports whether the source type is one of the supported values.
func (s SourceType) Valid() bool {
	switch s {
	case SourceGame, SourceServer, SourcePayment:
		return true
	default:
		return false
	}
}
