package dict

import (
	"sync"
)

type (
	// SafeDict -
	SafeDict[K any, V any] interface {
		RLock(func(DictR[K, V]))
		WLock(func(Dict[K, V]))
	}

	// SafeSet -
	SafeSet[K any] interface {
		RLock(func(SetR[K]))
		WLock(func(Set[K]))
	}
)

// NewSafeDict -
func NewSafeDict[K any, V any](d Dict[K, V]) *safeDict[K, V] {
	return &safeDict[K, V]{d: d}
}

// NewSafeSet -
func NewSafeSet[K any](s Set[K]) *safeSet[K] {
	return &safeSet[K]{s: s}
}

type (
	safeDict[K any, V any] struct {
		d  Dict[K, V]
		mx sync.RWMutex
	}
	safeSet[K any] struct {
		s  Set[K]
		mx sync.RWMutex
	}
)

// RLock -
func (sd *safeDict[K, V]) RLock(f func(DictR[K, V])) {
	sd.mx.RLock()
	defer sd.mx.RUnlock()
	f(sd.d)
}

// WLock -
func (sd *safeDict[K, V]) WLock(f func(Dict[K, V])) {
	sd.mx.Lock()
	defer sd.mx.Unlock()
	f(sd.d)
}

// RLock -
func (ss *safeSet[K]) RLock(f func(SetR[K])) {
	ss.mx.RLock()
	defer ss.mx.RUnlock()
	f(ss.s)
}

// WLock -
func (ss *safeSet[K]) WLock(f func(Set[K])) {
	ss.mx.Lock()
	defer ss.mx.Unlock()
	f(ss.s)
}
