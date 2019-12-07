package ffsm

import "context"

type ctxKey uint

const (
	dispatcherCtxKey     ctxKey = 1
	sourceStateCtxKey    ctxKey = 2
	distanateStateCtxKey ctxKey = 3
)

func hydrateContextForAction(ctx context.Context, dispatcher Dispatcher, src, dst State) context.Context {
	ctx = context.WithValue(ctx, dispatcherCtxKey, dispatcher)
	ctx = context.WithValue(ctx, sourceStateCtxKey, src)
	return context.WithValue(ctx, distanateStateCtxKey, dst)
}

// GetFSMDispatcher returns dispatcher of FSM from context.
func GetFSMDispatcher(ctx context.Context) Dispatcher {
	return ctx.Value(dispatcherCtxKey).(Dispatcher)
}

// GetSrcState returns source state from context.
func GetSrcState(ctx context.Context) State {
	return ctx.Value(sourceStateCtxKey).(State)
}

// GetDstState returns destinate state from context.
func GetDstState(ctx context.Context) State {
	return ctx.Value(distanateStateCtxKey).(State)
}
