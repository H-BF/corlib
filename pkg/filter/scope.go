package filter

type (
	//Scope scope interface
	Scope interface {
		privateScope()
	}

	// ScopedNot -
	ScopedNot struct {
		Scope
	}

	// ScopedAnd -
	ScopedAnd struct {
		L, R Scope
	}

	// ScopedAll -
	ScopedAll struct {
		Scs []Scope
	}

	// ScopedAny -
	ScopedAny struct {
		Scs []Scope
	}

	// ScopedOr -
	ScopedOr struct {
		L, R Scope
	}

	// NoScope -
	NoScope struct{}

	scope4FuncAny[T any, F ~func(T) bool] [1]F
)

func (ScopedNot) privateScope()           {}
func (ScopedOr) privateScope()            {}
func (ScopedAnd) privateScope()           {}
func (NoScope) privateScope()             {}
func (ScopedAll) privateScope()           {}
func (ScopedAny) privateScope()           {}
func (scope4FuncAny[T, F]) privateScope() {}

// Test -
func (f scope4FuncAny[T, F]) Test(arg T) bool {
	return f[0](arg)
}

// ScopeFromFunc -
func ScopeFromFunc[T any, F ~func(T) bool](f F) scope4FuncAny[T, F] {
	return scope4FuncAny[T, F]{f}
}

// And logical and cope
func And(t1 Scope, t2 Scope) Scope {
	return ScopedAnd{L: t1, R: t2}
}

// Or logical or scope
func Or(t1 Scope, t2 Scope) Scope {
	return ScopedOr{L: t1, R: t2}
}

// All -
func All(sc ...Scope) Scope {
	return ScopedAll{Scs: sc}
}

// Any -
func Any(sc ...Scope) Scope {
	return ScopedAny{Scs: sc}
}

// Not negate scope
func Not(t Scope) Scope {
	return ScopedNot{Scope: t}
}
