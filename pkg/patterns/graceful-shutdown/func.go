package graceful_shutdown

import (
	"context"

	"github.com/H-BF/corlib/pkg/functional"
)

// Func makes task with no output
func Func(f func(ctx context.Context)) func0 { //nolint:revive
	return func0{
		funcTask: funcTask{
			from: functional.MustCallableOf(f),
		},
	}
}

// Func1 makes task with 1 val output
func Func1[T1 any](f func(ctx context.Context) T1) func1[T1] { //nolint:revive
	return func1[T1]{
		funcTask: funcTask{
			from: functional.MustCallableOf(f),
		},
	}
}

// Func2 makes task with 2 values output
func Func2[T1 any, T2 any](f func(ctx context.Context) (T1, T2)) func2[T1, T2] { //nolint:revive
	return func2[T1, T2]{
		funcTask: funcTask{
			from: functional.MustCallableOf(f),
		},
	}
}

// Func3 makes task with 3 values output
func Func3[T1 any, T2 any, T3 any](f func(ctx context.Context) (T1, T2, T3)) func3[T1, T2, T3] { //nolint:revive
	return func3[T1, T2, T3]{
		funcTask: funcTask{
			from: functional.MustCallableOf(f),
		},
	}
}

// Func4 makes task with 4 values output
func Func4[T1 any, T2 any, T3 any, T4 any](f func(ctx context.Context) (T1, T2, T3, T4)) func4[T1, T2, T3, T4] { //nolint:revive
	return func4[T1, T2, T3, T4]{
		funcTask: funcTask{
			from: functional.MustCallableOf(f),
		},
	}
}

// Func5 makes task with 5 values output
func Func5[T1 any, T2 any, T3 any, T4 any, T5 any](f func(ctx context.Context) (T1, T2, T3, T4, T5)) func5[T1, T2, T3, T4, T5] { //nolint:revive
	return func5[T1, T2, T3, T4, T5]{
		funcTask: funcTask{
			from: functional.MustCallableOf(f),
		},
	}
}

//
//                              --== IMPL ==--
//

type (
	funcTask struct {
		from functional.Callable
		into functional.Callable
	}

	func0 struct {
		funcTask
	}
	func1[T1 any] struct {
		funcTask
	}
	func2[T1 any, T2 any] struct {
		funcTask
	}
	func3[T1 any, T2 any, T3 any] struct {
		funcTask
	}
	func4[T1 any, T2 any, T3 any, T4 any] struct {
		funcTask
	}
	func5[T1 any, T2 any, T3 any, T4 any, T5 any] struct {
		funcTask
	}
)

func (fn funcTask) consume(args ...interface{}) error {
	if fn.into != nil {
		return fn.into.InvokeNoResult(args...)
	}
	return nil
}

// Exec impl 'Task' interface
func (fn funcTask) Exec(ctx context.Context) TaskCompletion {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		rets, e := fn.from.Invoke(ctx)
		if e == nil {
			args := append([]interface{}{ctx}, rets...)
			e = fn.consume(args...)
		}
		if e != nil {
			panic(e)
		}
	}()
	return ch
}

// Consume it sets result consumer func
func (fn func1[T1]) Consume(f func(context.Context, T1)) func1[T1] {
	fn.into = functional.MustCallableOf(f)
	return fn
}

// Consume it sets result consumer func
func (fn func2[T1, T2]) Consume(f func(context.Context, T1, T2)) func2[T1, T2] {
	fn.into = functional.MustCallableOf(f)
	return fn
}

// Consume it sets result consumer func
func (fn func3[T1, T2, T3]) Consume(f func(context.Context, T1, T2, T3)) func3[T1, T2, T3] {
	fn.into = functional.MustCallableOf(f)
	return fn
}

// Consume it sets result consumer func
func (fn func4[T1, T2, T3, T4]) Consume(f func(context.Context, T1, T2, T3, T4)) func4[T1, T2, T3, T4] {
	fn.into = functional.MustCallableOf(f)
	return fn
}

// Consume it sets result consumer func
func (fn func5[T1, T2, T3, T4, T5]) Consume(f func(context.Context, T1, T2, T3, T4, T5)) func5[T1, T2, T3, T4, T5] {
	fn.into = functional.MustCallableOf(f)
	return fn
}
