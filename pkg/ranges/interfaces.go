package ranges

import (
	"fmt"
	"reflect"
)

type (
	boundTagID interface{ privateTagID() }

	Upper struct{ boundTagID }

	Lower struct{ boundTagID }

	BoundTag interface {
		boundTagID
		Lower | Upper
	}

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

	Factory[T any] interface {
		MinMax() (min T, max T)
		Bound(tag boundTagID, v T, exclude bool) Bound[T]
		Range(lower T, lExcluded bool, upper T, uExcluded bool) Range[T]
	}
)
