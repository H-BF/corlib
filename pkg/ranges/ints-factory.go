package ranges

import (
	"reflect"
)

// IntsFactory 'integers' class ranges factory constructor
func IntsFactory[T Ints](_ T) Factory[T] {
	return intsFactory[T]{}
}

type intsFactory[T Ints] struct{}

var _ Factory[int] = (*intsFactory[int])(nil)

func (f intsFactory[T]) Bound(tag boundTagID, val T, exclude bool) Bound[T] {
	var ret Bound[T]
	switch reflect.Indirect(reflect.ValueOf(tag)).Interface().(type) {
	case Upper:
		ret = new(intsBound[T, Upper])
	case Lower:
		ret = new(intsBound[T, Lower])
	default:
		panic("unexpected behavior reached")
	}
	ret.SetValue(val, exclude)
	return ret
}

func (f intsFactory[T]) Range(lower T, lExcluded bool, upper T, uExcluded bool) Range[T] {
	ret := new(intsRange[T])
	ret.lower.SetValue(lower, lExcluded)
	ret.upper.SetValue(upper, uExcluded)
	return ret
}

func (f intsFactory[T]) MinMax() (mi T, ma T) {
	if a := ^T(0); a > 0 {
		ma = a
	} else {
		bits := reflect.TypeOf(a).Bits()
		mi = ^T(1) << (bits - 2)
		ma = ^mi
	}
	return
}
