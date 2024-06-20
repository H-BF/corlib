package tm

import (
	"context"
)

type taskMangerCtxKey struct{}

// TaskManagerFromContext ...
func TaskManagerFromContext(ctx context.Context) TaskManger {
	p, _ := ctx.Value(taskMangerCtxKey{}).(TaskManger)
	return p
}

// TaskManagerToContext ...
func TaskManagerToContext(ctx context.Context, tm TaskManger) context.Context {
	return context.WithValue(ctx, taskMangerCtxKey{}, tm)
}
