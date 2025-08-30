package option

// Match -
func Match[T any](v ValueOf[T]) optMatcher[T] {
	return optMatcher[T](v)
}

type (
	optMatcher[T any] ValueOf[T]
	optNone[T any]    optMatcher[T]
	optSome[T any]    optMatcher[T]
)

// Some -
func (x optMatcher[T]) Some(f func(T)) optNone[T] {
	a := (*ValueOf[T])(&x)
	if obj, ok := a.Maybe(); ok {
		f(obj)
	}
	return optNone[T](x)
}

// None -
func (x optMatcher[T]) None(f func()) optSome[T] {
	if a := (*ValueOf[T])(&x); a.IsNone() {
		f()
	}
	return optSome[T](x)
}

// Some -
func (x optSome[T]) Some(f func(T)) {
	_ = optMatcher[T](x).Some(f)
}

// None -
func (x optNone[T]) None(f func()) {
	_ = optMatcher[T](x).None(f)
}
