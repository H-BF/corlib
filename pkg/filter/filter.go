package filter

import (
	"fmt"
	"reflect"
)

// SimpleFilter -
type SimpleFilter[T any] func(T) bool

// InitFromScope -
func (ft *SimpleFilter[T]) InitFromScope(sc Scope) error {
	f, e := ft.fromScope(sc)
	if e == nil {
		*ft = f
	}
	return e
}

func (ft SimpleFilter[T]) allOrAnyScope(all bool, scs ...Scope) (ret func(T) bool, err error) {
	var fns []func(T) bool
	for _, sc := range scs {
		var f func(T) bool
		if f, err = ft.fromScope(sc); err != nil {
			return nil, err
		}
		fns = append(fns, f)
	}
	ret = func(arg T) bool {
		r := false
		for _, f := range fns {
			r = f(arg)
			if (!r && all) || (r && !all) {
				break
			}
		}
		return r
	}
	return ret, nil
}

func (ft SimpleFilter[T]) fromScope(sc Scope) (ret func(T) bool, err error) {
	switch t := sc.(type) {
	case ScopedAnd:
		ret, err = ft.allOrAnyScope(true, t.L, t.R)
	case ScopedOr:
		ret, err = ft.allOrAnyScope(false, t.L, t.R)
	case ScopedNot:
		var f func(T) bool
		if f, err = ft.fromScope(t.Scope); err == nil {
			ret = func(arg T) bool {
				return !f(arg)
			}
		}
	case ScopedAll:
		ret, err = ft.allOrAnyScope(true, t.Scs...)
	case ScopedAny:
		ret, err = ft.allOrAnyScope(false, t.Scs...)
	case NoScope:
		ret = func(_ T) bool {
			return true
		}
	default:
		if v := reflect.ValueOf(t); v.IsValid() {
			dest := reflect.ValueOf(&ret).Elem()
			destType := dest.Type()
			n := v.NumMethod()
			for i := 0; ret == nil && i < n; i++ {
				meth := v.Method(i)
				methType := meth.Type()
				if methType.AssignableTo(destType) {
					dest.Set(meth)
				}
			}
		}
		if ret == nil {
			var a T
			err = fmt.Errorf(
				"scope '%T' has no any method to apply to '%T' type value",
				t, a,
			)
		}
	}
	return ret, err
}
