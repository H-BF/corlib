package lazy

import (
	"sync"
	"sync/atomic"
)

//OnceInitValue переменная которая может быть проиинициализировано только один раз
type OnceInitValue interface {
	GetOk() (interface{}, bool)
	Get() interface{}
	GetOrDefault() (interface{}, interface{})
	Assign(interface{})
}

//jOnceInitValue
type onceInitValueImpl struct {
	atomic.Value
}

type onceInitValueSync struct {
	sync.Once
}

//MakeOnceInitValue makes once init value holder
func MakeOnceInitValue() OnceInitValue {
	ret, initializer := new(onceInitValueImpl), new(onceInitValueSync)
	ret.Store(func() interface{} {
		return initializer
	})
	return ret
}

//GetOk ...
func (ff *onceInitValueImpl) GetOk() (ret interface{}, ok bool) {
	switch t := ff.Load().(type) {
	case func() interface{}:
		switch t1 := t().(type) {
		case *onceInitValueSync:
		default:
			ret, ok = t1, true
		}
	}
	return
}

//Get ...
func (ff *onceInitValueImpl) Get() (ret interface{}) {
	ret, _ = ff.GetOk()
	return
}

//GetOrDefault ...
func (ff *onceInitValueImpl) GetOrDefault() (ret interface{}, defaultVal interface{}) {
	var ok bool
	if ret, ok = ff.GetOk(); !ok {
		ret = defaultVal
	}
	return
}

//Assign ...
func (ff *onceInitValueImpl) Assign(value interface{}) {
	switch t := ff.Load().(type) {
	case func() interface{}:
		switch t1 := t().(type) {
		case *onceInitValueSync:
			t1.Do(func() {
				ff.Store(func() interface{} {
					return value
				})
			})
		}
	}
}
