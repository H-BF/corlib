package graceful_shutdown

import (
	"context"
)

// Chan it makes task from chan
func Chan[T any](c <-chan T) *sChan[T] { //nolint:revive
	return &sChan[T]{c: c}
}

type sChan[T any] struct {
	c       <-chan T
	into    func(context.Context, T)
	ifNoRes func() T
}

// IfNoResult it sets func to calc value if unable to get value ftm chan
func (c *sChan[T]) IfNoResult(f func() T) *sChan[T] {
	c.ifNoRes = f
	return c
}

// Consume it sets consumer func
func (c *sChan[T]) Consume(f func(context.Context, T)) *sChan[T] {
	c.into = f
	return c
}

func (c *sChan[T]) consume(ctx context.Context, v T) {
	if c.into != nil {
		c.into(ctx, v)
	}
}

// Exec impl 'Task' interface
func (c *sChan[T]) Exec(ctx context.Context) TaskCompletion {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		select {
		case <-ctx.Done():
		case ret, ok := <-c.c:
			if ok {
				c.consume(ctx, ret)
			} else if c.ifNoRes != nil && c.into != nil {
				c.consume(ctx, c.ifNoRes())
			}
		}
	}()
	return ch
}
