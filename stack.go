package ffsm

import "context"

// Stack actions of transition.
type Stack map[StackKey][]Procedure

// StackKey is the identifier of the transition.
type StackKey struct {
	Src State
	Dst State
}

// Add registration action.
func (r Stack) Add(src State, dst State, p ...Procedure) Stack {
	if r == nil {
		panic("Stack.Add: stack is empty")
	}

	e := StackKey{Src: src, Dst: dst}
	if r[e] == nil {
		r[e] = []Procedure{}
	}
	r[e] = append(r[e], p...)

	return r
}

// Get return actions of event.
func (r Stack) Get(src, dst State) []Procedure {
	if r == nil {
		panic("Stack.Get: stack is empty")
	}
	return r[StackKey{Src: src, Dst: dst}]
}

// Procedure handler of transition.
type Procedure func(ctx context.Context) (context.Context, error)
