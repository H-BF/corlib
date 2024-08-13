package lazy

import (
	"sync"
	"sync/atomic"
)

// ValOrErr -
type ValOrErr[T any] struct {
	Val T
	Err error
}

// Initializer holds effective lazy init algorithm
type Initializer[T any] interface {
	Value() T
}

// MakeInitializer get effective lazy init algorithm
func MakeInitializer[T any](initializer func() T) Initializer[T] {
	type fetcherFunc = func() T
	var holder atomic.Value
	var once sync.Once
	var value T
	holder.Store(fetcherFunc(func() T {
		once.Do(func() {
			defer holder.Store(fetcherFunc(func() T {
				return value
			}))
			value = initializer()
		})
		return value
	}))
	return initializerImpl[T](func() T {
		return holder.Load().(fetcherFunc)()
	})
}

type initializerImpl[T any] func() T

// Value -
func (impl initializerImpl[T]) Value() T {
	return impl()
}
