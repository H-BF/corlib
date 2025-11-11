package result

// Wrap -
func Wrap[T any](v T, e error) (ret Value[T]) {
	ret.V, ret.E = v, e
	return ret
}

// Value -
type Value[T any] struct {
	V T
	E error
}

type valueFail struct {
	e error
}
type valueOk[T any] struct {
	v      T
	actual bool
}

// Set -
func (v *Value[T]) Set(val T, e error) {
	v.V, v.E = val, e
}

// Val -
func (v Value[T]) Val() (T, error) {
	var zero T
	if v.E != nil {
		return zero, v.E
	}
	return v.V, nil
}

// Ok -
func (v Value[T]) Ok(f func(T)) valueFail {
	if v.E == nil {
		f(v.V)
	}
	return valueFail{e: v.E}
}

// Fail -
func (v Value[T]) Fail(f func(error)) valueOk[T] {
	if v.E != nil {
		f(v.E)
	}
	return valueOk[T]{
		v:      v.V,
		actual: v.E == nil,
	}
}

// Fail -
func (v valueFail) Fail(f func(error)) {
	if v.e != nil {
		f(v.e)
	}
}

// Ok -
func (v valueOk[T]) Ok(f func(T)) {
	if v.actual {
		f(v.v)
	}
}
