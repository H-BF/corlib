package ranges

import (
	"bytes"
	"fmt"
	"sort"
)

// MultiRange multi range def
type MultiRange[T any] struct {
	factory Factory[T]
	ranges  []Range[T]
}

// NewMultiRange creates new multi range
func NewMultiRange[T any](f Factory[T]) MultiRange[T] {
	return MultiRange[T]{factory: f}
}

// Iterate iterates ranges from multi range
func (mr MultiRange[T]) Iterate(f func(r Range[T]) bool) {
	for i := range mr.ranges {
		if !f(mr.ranges[i].Copy()) {
			break
		}
	}
}

// Update updates ranges in multi range with some strategies
func (mr *MultiRange[T]) Update(
	strategy CombineStrategy,
	rr ...Range[T],
) {
	if len(rr) == 0 {
		return
	}
	values := append(
		append(
			make([]Range[T], 0, len(rr)+len(mr.ranges)),
			mr.ranges...),
		rr...)
	ret := values[:0]
	CombineRanges(
		func(l, u Bound[T]) Range[T] {
			var v T
			r := mr.factory.Range(v, false, v, true)
			r.SetBounds(l, u)
			return r
		},
		func(v Range[T]) bool {
			ret = append(ret, v)
			return true
		}, strategy, values...)
	mr.ranges = ret
}

// String impl fmt.Stringer
func (mr MultiRange[T]) String() string {
	b := bytes.NewBuffer(nil)
	mr.Iterate(func(r Range[T]) bool {
		_, _ = fmt.Fprintf(b, "%s", r)
		return true
	})
	return b.String()
}

// Len count of ranges in multi range
func (mr MultiRange[T]) Len() int {
	return len(mr.ranges)
}

// At get copy of range at pos
func (mr MultiRange[T]) At(i int) Range[T] {
	if !(i >= 0 && i < len(mr.ranges)) {
		return nil
	}
	return mr.ranges[i].Copy()
}

// Search searches range where 'v' is in
func (mr MultiRange[T]) Search(v T) (int, bool) {
	x := mr.factory.Bound(Lower{}, v, false)
	n := sort.Search(len(mr.ranges), func(i int) bool {
		l, u := mr.ranges[i].Bounds()
		diff1, diff2 := u.Cmp(x), l.Cmp(x)
		return !(diff1 < 0 && diff2 < 0)
	})
	return n, n >= 0 && n < len(mr.ranges) && mr.ranges[n].Contains(v)
}
