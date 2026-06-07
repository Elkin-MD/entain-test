package enum

// State is the effect a transaction has on a balance.
type State string

const (
	StateWin  State = "win"
	StateLose State = "lose"
)

// Valid reports whether the state is one of the supported values.
func (s State) Valid() bool {
	return s == StateWin || s == StateLose
}
