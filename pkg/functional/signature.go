package functional

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/pkg/errors"
)

//SignatureKey comparable key of signature
type SignatureKey = interface{}

//Signature is a function signature
type Signature interface {
	EqualTo(Signature) bool
	FromOutputValues() Signature
	ArgsInfo() (args []reflect.Type, variadic bool)
	Key() SignatureKey
}

//MustSignatureOf inspect function signature or panic if error
func MustSignatureOf(funcType interface{}) Signature {
	s, e := MaySignatureOf(funcType)
	if e != nil {
		panic(e)
	}
	return s
}

//MaySignatureOf inspect function signature or return error
func MaySignatureOf(funcType interface{}) (Signature, error) {
	result := &signature{}
	v := reflect.Indirect(reflect.ValueOf(funcType))
	t := v.Type()
	if t.Kind() != reflect.Func {
		return nil, errors.New("MaySignatureOf: function type value requires")
	}
	result.variadic = t.IsVariadic()
	nIn, nOut := t.NumIn(), t.NumOut()
	for i := 0; i < nIn; i++ {
		tA := t.In(i)
		if (i+1) == nIn && result.variadic {
			tA = tA.Elem()
		}
		result.argsIn = append(result.argsIn, tA)
	}
	for i := 0; i < nOut; i++ {
		result.argsOut = append(result.argsOut, t.Out(i))
	}
	var o sync.Once
	var key SignatureKey
	getter := &result.keyGetter
	in, variadic := result.argsIn, result.variadic
	result.keyGetter = func() SignatureKey {
		o.Do(func() {
			key = makeSignatureKey(in, variadic)
			*getter = func() SignatureKey {
				return key
			}
		})
		return key
	}
	return result, nil
}

type signature struct {
	argsIn    []reflect.Type
	argsOut   []reflect.Type
	variadic  bool
	keyGetter func() SignatureKey
}

//ArgsInfo input arguments info
func (si *signature) ArgsInfo() (args []reflect.Type, variadic bool) {
	return si.argsIn, si.variadic
}

//EqualTo check if signature is equal to the other
func (si *signature) EqualTo(other Signature) bool {
	argsR, variadic := other.ArgsInfo()
	argsL := si.argsIn
	if len(argsL) == len(argsR) {
		for i := range argsL {
			if argsL[i] != argsR[i] {
				return false
			}
		}
		return si.variadic == variadic
	}
	return false
}

//FromOutputValues make signature from output data
func (si *signature) FromOutputValues() Signature {
	result := &signature{argsIn: si.argsOut}
	var o sync.Once
	var key SignatureKey
	getter := &result.keyGetter
	in, variadic := result.argsIn, result.variadic
	result.keyGetter = func() SignatureKey {
		o.Do(func() {
			key = makeSignatureKey(in, variadic)
			*getter = func() SignatureKey {
				return key
			}
		})
		return key
	}
	return result
}

//Key makes signature comparable key / it mau be used in map(s)
func (si *signature) Key() SignatureKey {
	return si.keyGetter()
}

func makeSignatureKey(args []reflect.Type, variadic bool) SignatureKey {
	fields := make([]reflect.StructField, 0, len(args)+1)
	ty := reflect.TypeOf((*interface{})(nil)).Elem()
	for i := range args {
		fields = append(fields, reflect.StructField{
			Type: ty,
			Name: fmt.Sprintf("F%v", i),
		})
	}
	fields = append(fields, reflect.StructField{
		Type: reflect.TypeOf((*bool)(nil)).Elem(),
		Name: fmt.Sprintf("F%v", len(args)),
	})
	st := reflect.New(reflect.StructOf(fields)).Elem()
	for i := range args {
		st.Field(i).Set(reflect.ValueOf(args[i]))
	}
	st.Field(len(args)).SetBool(variadic)
	return st.Interface()
}
