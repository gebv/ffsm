package ffsm

// AnyState is the any state.
const AnyState State = "*"

// UnknownState is the unknown state.
const UnknownState State = ""

// State name of state.
type State string

// Match return true for equal state.
func (s State) Match(in State) bool {
	return s == in
}

// String returns string name of state.
func (s State) String() string {
	return string(s)
}
