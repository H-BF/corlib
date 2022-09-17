package ranges

import (
	"fmt"
	"reflect"
)

type (
	boundTagID interface{ privateTagID() }

	//Upper is constraint of upper bound
	Upper struct{ boundTagID }

	//Lower is constraint of lower bound
	Lower struct{ boundTagID }

	//BoundTag bound kind constraint
	BoundTag interface {
		boundTagID
		Lower | Upper
	}

	//Bound an interface of range bound
	Bound[T any] interface {
		fmt.Stringer
		GetValue() (val T, excluded bool)
		SetValue(val T, excluded bool)
		Is(boundTagID) bool
		Adjacent() Bound[T]
		Type() reflect.Type
		Cmp(other Bound[T]) int
		IsIn(v Range[T]) bool
		Copy() Bound[T]
		AsIncluded() Bound[T]
		AsExcluded() Bound[T]
	}

	//Range an interface of range
	Range[T any] interface {
		fmt.Stringer
		IsNull() bool
		Normalize() Range[T]
		Bounds() (lower, upper Bound[T])
		SetBounds(lower, upper Bound[T])
		Filter(f func(x T, included bool) bool, data ...T)
		Contains(v T) bool
		Copy() Range[T]
	}

	//Factory ranges factory
	Factory[T any] interface {
		MinMax() (min T, max T)
		Bound(tag boundTagID, v T, exclude bool) Bound[T]
		Range(lower T, lExcluded bool, upper T, uExcluded bool) Range[T]
	}
)
