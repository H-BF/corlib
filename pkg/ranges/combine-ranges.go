package ranges

import (
	"sort"
)

// CombineStrategy combine strategy ID
type CombineStrategy = uint8

const (
	// CombineMerge merge strategy
	CombineMerge CombineStrategy = iota

	// CombineExclude exclude strategy
	CombineExclude
)

// CombineRanges combines set of ganges with one of strategy
func CombineRanges[T any](
	rangeConstructor func(l, u Bound[T]) Range[T],
	consume func(Range[T]) bool,
	strategy CombineStrategy,
	ranges ...Range[T],
) {
	pts := make([]Bound[T], 0, len(ranges)*2)
	for _, v := range ranges {
		if !v.IsNull() {
			a, b := v.Bounds()
			pts = append(pts, a, b)
		}
	}
	sort.Slice(pts, func(i, j int) bool {
		l, r := pts[i], pts[j]
		n := l.Cmp(r)
		if n == 0 && l.Is(Lower{}) && r.Is(Upper{}) {
			n = -1
		}
		return n < 0
	})
	switch strategy {
	case CombineMerge:
		combineMerge(rangeConstructor, consume, pts)
	case CombineExclude:
		combineExclude(rangeConstructor, consume, pts)
	default:
		panic("unexpected strategy")
	}
}

func combineMerge[T any](
	rangeConstructor func(l, u Bound[T]) Range[T],
	consume func(Range[T]) bool,
	sortedBounds []Bound[T],
) {
	var lwr Bound[T]
	lvl := 0
	for i := range sortedBounds {
		if p := sortedBounds[i]; p.Is(Lower{}) {
			lvl++
			if lvl == 1 {
				lwr = p
			}
		} else if lvl--; lvl == 0 {
			v := rangeConstructor(lwr, p)
			if v = v.Normalize(); !v.IsNull() && !consume(v) {
				return
			}
		}
		if lvl < 0 {
			lvl = 0
		}
	}
}

func combineExclude[T any](
	rangeConstructor func(l, u Bound[T]) Range[T],
	consume func(Range[T]) bool,
	SortedBounds []Bound[T],
) {
	n := 0
	for i := range SortedBounds {
		var v Range[T]
		if b := SortedBounds[i]; b.Is(Lower{}) {
			if n++; n == 2 {
				v = rangeConstructor(SortedBounds[i-1], b)
			}
		} else if n--; n == 0 {
			v = rangeConstructor(SortedBounds[i-1], b)
		}
		if n < 0 {
			n = 0
		}
		if v == nil {
			continue
		}
		if v = v.Normalize(); !v.IsNull() && !consume(v) {
			return
		}
	}
}
