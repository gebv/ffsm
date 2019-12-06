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

func (s State) String() string {
	return string(s)
}

// Set to set new value.
func (s *State) Set(newState State) {
	*s = newState
}
