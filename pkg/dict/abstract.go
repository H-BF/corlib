package dict

import (
	"slices"
)

type (
	// DictR abstact dictionary readonly interface
	DictR[Tk any, Tv any] interface {
		Len() int
		Get(k Tk) (v Tv, ok bool)
		Keys() []Tk
		Items() Items[Tk, Tv]
		Iterate(f func(k Tk, v Tv) bool)
		At(k Tk) Tv
		Eq(other DictR[Tk, Tv], valuesEq func(vL, vR Tv) bool) bool
	}

	// Dict abstact interface
	Dict[Tk any, Tv any] interface {
		DictR[Tk, Tv]
		Clear()
		Del(keys ...Tk)
		Put(k Tk, v Tv)
		PutMany(...KV[Tk, Tv])
		Insert(k Tk, v Tv) bool
	}

	// KV -
	KV[K any, V any] struct {
		K K
		V V
	}

	// Items -
	Items[K any, V any] []KV[K, V]

	// SetR abstract set readonly interface
	SetR[T any] interface {
		Len() int
		Contains(k T) bool
		ContainsAny(k ...T) bool
		Iterate(f func(k T) bool)
		Values() []T
		Eq(SetR[T]) bool
	}

	// Set abstract set interface
	Set[T any] interface {
		SetR[T]
		Clear()
		Del(keys ...T)
		Put(k T)
		PutMany(vals ...T)
		Insert(k T) bool
	}
)

// Reserve -
func (i *Items[K, V]) Reserve(n int) *Items[K, V] {
	*i = slices.Grow(*i, n)
	return i
}

// Add -
func (i *Items[K, V]) Add(k K, v V) *Items[K, V] {
	*i = append(*i, KV[K, V]{k, v})
	return i
}
