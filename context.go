package ffsm

import "context"

type ctxKey uint

const (
	sourceStateCtxKey    ctxKey = 2
	distanateStateCtxKey ctxKey = 3
)

func hydrateContextForAction(ctx context.Context, src, dst string) context.Context {
	ctx = context.WithValue(ctx, sourceStateCtxKey, src)
	return context.WithValue(ctx, distanateStateCtxKey, dst)
}

// GetSrcState returns source state from context.
func GetSrcState(ctx context.Context) string {
	return ctx.Value(sourceStateCtxKey).(string)
}

// GetDstState returns destinate state from context.
func GetDstState(ctx context.Context) string {
	return ctx.Value(distanateStateCtxKey).(string)
}
