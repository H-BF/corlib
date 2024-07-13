package plain_config

import (
	"reflect"

	"github.com/H-BF/corlib/pkg/functional"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"go.uber.org/multierr"
)

// RegisterTypeCast -
func RegisterTypeCast[T any](f func(any) (T, error)) {
	regTypeCastFunc(typeCastFunc[T](f), false)
}

func typeCast[T any](in any, ret *T) error {
	var c typeCastFunc[T]
	err := c.load()
	if err == nil {
		*ret, err = c(in)
	}
	return err
}

type typeCastFunc[T any] func(any) (T, error)

func (f *typeCastFunc[T]) load() error {
	tyDest := reflect.TypeOf((*T)(nil)).Elem()
	kindDest := tyDest.Kind()
	if kindDest == reflect.Interface {
		*f = func(in any) (r T, e error) {
			reflect.ValueOf(&r).Elem().
				Set(reflect.ValueOf(in))
			return
		}
		return nil
	}
	*f = func(in any) (T, error) {
		var ret T
		if castInvoker := typeCastInvokers[tyDest]; castInvoker != nil {
			var v interface{}
			var e error
			if e1 := castInvoker.InvokeNoResult(in, &v, &e); e1 != nil || e != nil {
				return ret, multierr.Combine(e1, e)
			}
			reflect.ValueOf(&ret).Elem().
				Set(
					reflect.ValueOf(v),
				)
		} else {
			x := reflect.ValueOf(in)
			if !x.Type().ConvertibleTo(tyDest) {
				return ret, errors.WithMessagef(ErrTypeCastNotSupported, "for-type('%s')", tyDest)
			}
			reflect.ValueOf(&ret).Elem().
				Set(
					x.Convert(tyDest),
				)
		}
		return ret, nil
	}
	return nil
}

var (
	typeCastInvokers = make(map[reflect.Type]functional.Callable)
)

func constructTypeCastInvoker[T any](c typeCastFunc[T]) functional.Callable {
	return functional.MustCallableOf(
		func(in interface{}, ret *interface{}, err *error) {
			*ret, *err = c(in)
		},
	)
}

func regTypeCastFunc[T any](c typeCastFunc[T], failIfOverride bool) {
	var a *T
	ty := reflect.TypeOf(a).Elem()
	if failIfOverride && typeCastInvokers[ty] != nil {
		panic(errors.Errorf("('%v') type cast is always registered", ty))
	}
	typeCastInvokers[ty] = constructTypeCastInvoker(c)
}

func init() {
	regTypeCastFunc(cast.ToBoolE, true)

	regTypeCastFunc(cast.ToInt8E, true)
	regTypeCastFunc(cast.ToInt16E, true)
	regTypeCastFunc(cast.ToInt32E, true)
	regTypeCastFunc(cast.ToInt64E, true)

	regTypeCastFunc(cast.ToUint8E, true)
	regTypeCastFunc(cast.ToUint16E, true)
	regTypeCastFunc(cast.ToUint32E, true)
	regTypeCastFunc(cast.ToUint64E, true)

	regTypeCastFunc(cast.ToIntE, true)
	regTypeCastFunc(cast.ToUintE, true)

	regTypeCastFunc(cast.ToStringE, true)

	regTypeCastFunc(cast.ToFloat32E, true)
	regTypeCastFunc(cast.ToFloat64E, true)

	regTypeCastFunc(cast.ToDurationE, true)
	regTypeCastFunc(cast.ToTimeE, true)

	regTypeCastFunc(typeCastNetCIDR, true)
	regTypeCastFunc(typeCastNetCIDRSlice, true)

	regTypeCastFunc(typeCastIP, true)
	regTypeCastFunc(typeCastIPSlice, true)

	regTypeCastFunc(typeCastSliceT[string], true)

	regTypeCastFunc(cast2uuid, true)
}
