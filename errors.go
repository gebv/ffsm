package ffsm

import "errors"

var (
	// ErrNotInitalState is the error returned by Machine when the is
	// not have initial state of Machine.
	ErrNotInitalState = errors.New("Is not set initial value of state")

	// ErrCtxCanceled is the error when context is canceled.
	ErrCtxCanceled = errors.New("Context canceled")

	// ErrNotRegTransition is the error returned by Machine from Dispatch method when the is
	// have not rules for current transition (src->dst not have actions).
	ErrNotRegTransition = errors.New("Not registred transition")
)

// DispatchError is the container with custom errors for dispatcher.
type DispatchError struct {
	ActionName        string
	SrcState          State
	DstState          State
	Err               error
	IsPanic           bool
	PanicStackRuntime string
}

func (e DispatchError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}
