package tm

import (
	"context"
)

// ContextWithCancel специфический контекст с канцеллером
type ContextWithCancel interface {
	context.Context
	Cancel()
}

// NewContextWithCancel делает специфический контекст с канцеллером
func NewContextWithCancel(ctx context.Context) ContextWithCancel {
	newCtx, canceller := context.WithCancel(ctx)
	return &contextWithCancel{Context: newCtx, cancel: canceller}
}

type contextWithCancel struct {
	context.Context
	cancel func()
}

func (cc *contextWithCancel) Cancel() {
	cc.cancel()
}
