package functional

import (
	"reflect"

	"github.com/pkg/errors"
)

// Callable is functional object interface
type Callable interface {
	Signature
	Invoke(...interface{}) ([]interface{}, error)
	InvokeNoResult(...interface{}) error
}

var (
	// ErrArgsNotMatched2Signature error when  arguments are not matched to signature
	ErrArgsNotMatched2Signature = errors.New("arguments are not matched to signature")
)

// MustInvoke call functional object or panic if error
func MustInvoke(c Callable, args ...interface{}) []interface{} {
	r, e := c.Invoke(args...)
	if e != nil {
		panic(e)
	}
	return r
}

// MustInvokeNoResult call functional object and ignore result or panic if error
func MustInvokeNoResult(c Callable, args ...interface{}) {
	if e := c.InvokeNoResult(args...); e != nil {
		panic(e)
	}
}

type callableImpl struct {
	Signature
	wrappedFunction reflect.Value
}

// MustCallableOf construct functional object or panic if error
func MustCallableOf(f interface{}) Callable {
	c, e := MayCallableOf(f)
	if e != nil {
		panic(e)
	}
	return c
}

// MayCallableOf construct functional object or return error
func MayCallableOf(funcObject interface{}) (Callable, error) {
	if c, ok := funcObject.(Callable); ok {
		return c, nil
	}
	var (
		ret callableImpl
		err error
	)
	if ret.Signature, err = MaySignatureOf(funcObject); err != nil {
		return nil, errors.Wrap(err, "MayCallableOf")
	}
	ret.wrappedFunction = reflect.Indirect(reflect.ValueOf(funcObject))
	return &ret, nil
}

func (obj *callableImpl) internalInvoke(ret *[]reflect.Value, args ...interface{}) (retErr error) {
	argsIn, variadic := obj.ArgsInfo()
	nMinArgs := len(argsIn)
	if variadic {
		nMinArgs--
	}
	nIn := len(args)
	if nIn < nMinArgs {
		return errors.Wrap(ErrArgsNotMatched2Signature, "not enough args")
	}
	if nIn > len(argsIn) && !variadic {
		return errors.Wrap(ErrArgsNotMatched2Signature, "too few args")
	}
	vargs := make([]reflect.Value, nIn)
	var checkArgType reflect.Type
	for i, arg := range args {
		if i < len(argsIn) {
			checkArgType = argsIn[i]
		}
		a := reflect.ValueOf(arg)
		if !a.IsValid() {
			a = reflect.New(checkArgType).Elem()
		}
		vargs[i] = a
	}
	defer func() {
		if r := recover(); r != nil {
			retErr = errors.Wrapf(ErrArgsNotMatched2Signature, "crashed: %v", r)
		}
	}()
	r := obj.wrappedFunction.Call(vargs)
	if ret != nil {
		*ret = r
	}
	return nil
}

// Invoke call functional object or return error
func (obj *callableImpl) Invoke(args ...interface{}) ([]interface{}, error) {
	const api = "callable/Invoke"
	var ret []reflect.Value
	if err := obj.internalInvoke(&ret, args...); err != nil {
		return nil, errors.Wrap(err, api)
	}
	if nRet := len(ret); nRet > 0 {
		result := make([]interface{}, nRet)
		for i := range ret {
			result[i] = ret[i].Interface()
		}
		return result, nil
	}
	return nil, nil
}

// InvokeNoResult call functional object and ignore result or return error
func (obj *callableImpl) InvokeNoResult(args ...interface{}) error {
	return obj.internalInvoke(nil, args...)
}
