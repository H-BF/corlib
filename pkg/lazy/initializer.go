package lazy

import (
	"sync"
	"sync/atomic"
)

// Initializer holds effective lazy init algorithm
type Initializer interface {
	Value() interface{}
}

// MakeInitializer get effective lazy init algorithm
func MakeInitializer(initializer func() interface{}) Initializer {
	type fetcherFunc func() interface{}
	var (
		once   sync.Once
		holder atomic.Value
		value  interface{}
		fetch  fetcherFunc = func() interface{} {
			return value
		}
		lazyFetch fetcherFunc = func() interface{} {
			once.Do(func() {
				if initializer != nil {
					value = initializer()
				}
				holder.Store(fetch)
			})
			return value
		}
	)
	holder.Store(lazyFetch)
	return &initializerImpl{
		valueGetter: func() (ret interface{}) {
			switch get := holder.Load().(type) {
			case fetcherFunc:
				ret = get()
			}
			return
		},
	}
}

type initializerImpl struct {
	valueGetter func() interface{}
}

func (impl *initializerImpl) Value() interface{} {
	return impl.valueGetter()
}
