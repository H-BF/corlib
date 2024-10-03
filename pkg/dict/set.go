package dict

type (
	//HSet hash based set
	HSet[T comparable] struct {
		impSet[T, hDictFactory[T, struct{}]]
	}

	//RBSet is the red-black tree based set
	RBSet[T any] struct {
		impSet[T, rbDictFactory[T, struct{}]]
	}

	factoryOfDict[Tk any, Tv any] interface {
		construct() Dict[Tk, Tv]
	}

	rbDictFactory[Tk any, Tv any] struct{} //nolint:unused

	hDictFactory[Tk comparable, Tv any] struct{} //nolint:unused
)

func (rbDictFactory[Tk, Tv]) construct() Dict[Tk, Tv] { //nolint:unused
	return new(RBDict[Tk, Tv])
}

func (hDictFactory[Tk, Tv]) construct() Dict[Tk, Tv] { //nolint:unused
	return new(HDict[Tk, Tv])
}

type impSet[T any, F factoryOfDict[T, struct{}]] struct {
	inner Dict[T, struct{}]
}

func (set *impSet[T, F]) init() {
	if set.inner == nil {
		var f F
		set.inner = f.construct()
	}
}

// Clear -
func (set *impSet[T, F]) Clear() {
	set.init()
	set.inner.Clear()
}

// Len -
func (set *impSet[T, F]) Len() int {
	set.init()
	return set.inner.Len()
}

// Del -
func (set *impSet[T, F]) Del(keys ...T) {
	set.init()
	set.inner.Del(keys...)
}

// Put -
func (set *impSet[T, F]) Put(k T) {
	set.init()
	set.inner.Put(k, struct{}{})
}

// PutMany -
func (set *impSet[T, F]) PutMany(vals ...T) {
	if len(vals) > 0 {
		set.init()
		for _, k := range vals {
			set.inner.Put(k, struct{}{})
		}
	}
}

// Insert -
func (set *impSet[T, F]) Insert(k T) bool {
	set.init()
	return set.inner.Insert(k, struct{}{})
}

// Contains -
func (set *impSet[T, F]) Contains(k T) bool {
	set.init()
	_, ok := set.inner.Get(k)
	return ok
}

// ContainsAny -
func (set *impSet[T, F]) ContainsAny(k ...T) (ok bool) {
	for _, v := range k {
		if ok = set.Contains(v); ok {
			break
		}
	}
	return ok
}

// Iterate -
func (set *impSet[T, F]) Iterate(f func(k T) bool) {
	set.init()
	set.inner.Iterate(func(k T, _ struct{}) bool {
		return f(k)
	})
}

// Values -
func (set *impSet[T, F]) Values() []T {
	ret := make([]T, 0, set.Len())
	set.Iterate(func(k T) bool {
		ret = append(ret, k)
		return true
	})
	return ret
}

// Eq -
func (set *impSet[T, F]) Eq(other Set[T]) bool {
	if set.Len() != other.Len() {
		return false
	}
	n := 0
	set.Iterate(func(k T) bool {
		ok := other.Contains(k)
		n += tern(ok, 1, 0)
		return ok
	})
	return n == set.Len()
}
