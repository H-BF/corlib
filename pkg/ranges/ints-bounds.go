package ranges

import (
	"bytes"
	"fmt"
	"reflect"
)

type intsBound[T Ints, Tag BoundTag] struct {
	Value    T
	Excluded bool
}

var (
	_ Bound[int] = (*intsBound[int, Upper])(nil)
	_ Bound[int] = (*intsBound[int, Lower])(nil)
)

func (b *intsBound[T, Tag]) cmpSameTag(x Bound[T]) int {
	r, xExcluded := x.GetValue()
	l := b.Value
	var ret int
	var delta T
	if l < r {
		delta = r - l
		ret = -1
	} else if l > r {
		delta = l - r
		ret = 1
	}
	if delta > 1 || (b.Excluded == xExcluded) {
		return ret
	}
	deltaEx := 1
	if b.Is(Lower{}) {
		if xExcluded {
			deltaEx = -1
		}
	} else if b.Excluded {
		deltaEx = -1
	}
	ret = int(delta)*ret + deltaEx
	switch {
	case ret < 0:
		return -1
	case ret > 0:
		return 1
	}
	return 0
}

func (b *intsBound[T, Tag]) Copy() Bound[T] {
	return &intsBound[T, Tag]{
		Excluded: b.Excluded,
		Value:    b.Value,
	}
}

func (b *intsBound[T, Tag]) GetValue() (val T, excluded bool) {
	val, excluded = b.Value, b.Excluded
	return
}

func (b *intsBound[T, Tag]) Is(id boundTagID) bool {
	var a Tag
	return reflect.ValueOf(a).Type() ==
		reflect.Indirect(reflect.ValueOf(id)).Type()
}

func (b *intsBound[T, Tag]) Cmp(x Bound[T]) int {
	var tag Tag
	if b.Is(tag) && x.Is(tag) {
		return b.cmpSameTag(x)
	}
	if b.Is(Upper{}) {
		return -x.Cmp(b)
	}
	var delta T
	var ret int
	xValue, xExcluded := x.GetValue()
	if b.Value < xValue {
		delta, ret = xValue-b.Value, -1
	} else if b.Value > xValue {
		delta, ret = b.Value-xValue, 1
	}
	if delta <= 2 {
		ret *= int(delta)
		if b.Excluded {
			ret++
		}
		if xExcluded {
			ret++
		}
	}
	switch {
	case ret < 0:
		return -1
	case ret > 0:
		return 1
	}
	return 0
}

func (b *intsBound[T, Tag]) SetValue(v T, excluded bool) {
	b.Value, b.Excluded = v, excluded
}

func (b *intsBound[T, Tag]) Adjacent() Bound[T] {
	if b.Is(Lower{}) {
		return &intsBound[T, Upper]{
			Value: b.Value, Excluded: !b.Excluded,
		}
	}
	return &intsBound[T, Lower]{
		Value: b.Value, Excluded: !b.Excluded,
	}
}

func (b *intsBound[T, Tag]) Type() reflect.Type {
	return reflect.TypeOf(b).Elem()
}

func (b *intsBound[T, Tag]) String() string {
	s := bytes.NewBuffer(nil)
	if b.Is(Upper{}) {
		_, _ = fmt.Fprintf(s, "%v", b.Value)
		if b.Excluded {
			_ = s.WriteByte(')')
		} else {
			_ = s.WriteByte(']')
		}
	} else {
		if b.Excluded {
			_ = s.WriteByte('(')
		} else {
			_ = s.WriteByte('[')
		}
		_, _ = fmt.Fprintf(s, "%v", b.Value)
	}
	return s.String()
}

func (b *intsBound[T, Tag]) IsIn(iv Range[T]) bool {
	l, u := iv.Bounds()
	return !iv.IsNull() &&
		b.Cmp(l) >= 0 &&
		b.Cmp(u) <= 0
}

func (b *intsBound[T, Tag]) AsIncluded() Bound[T] {
	if v, ex := b.GetValue(); ex {
		ok := false
		if b.Is(Lower{}) {
			v += 1
			ok = v > b.Value
		} else {
			v -= 1
			ok = v < b.Value
		}
		if ok {
			ret := b.Copy()
			ret.SetValue(v, false)
			return ret
		}
	}
	return b
}

func (b *intsBound[T, Tag]) AsExcluded() Bound[T] {
	if v, ex := b.GetValue(); !ex {
		ok := false
		if b.Is(Lower{}) {
			v -= 1
			ok = v < b.Value
		} else {
			v += 1
			ok = v > b.Value
		}
		if ok {
			ret := b.Copy()
			ret.SetValue(v, true)
			return ret
		}
	}
	return b
}
