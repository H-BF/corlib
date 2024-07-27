package visitor

type (
	visitorFunc[T any] func(T) error

	acceptorFunc[T any] func() T

	visitOneOf struct {
		vi   OneOf
		next *visitOneOf
	}
)

// Func2Acceptor wraps func as acceptor
func Func2Acceptor[T any](f func() T) Acceptor {
	return acceptorFunc[T](f)
}

// WrapAsAcceptor it makes simple acceptor from any data
func WrapAsAcceptor[T any](d T) Acceptor {
	return Func2Acceptor(func() T {
		return d
	})
}

func (f visitorFunc[T]) Visit(v interface{}) (err error) {
	switch t := v.(type) {
	case *T:
		if t != nil {
			err = f(*t)
		}
	case T:
		err = f(t)
	}
	return err
}

func (f visitorFunc[T]) visitOrNext(v interface{}, next Visitor) (err error) {
	switch t := v.(type) {
	case *T:
		if t != nil {
			err = f(*t)
		}
	case T:
		err = f(t)
	default:
		if next != nil {
			err = next.Visit(v)
		}
	}
	return err
}

func (f acceptorFunc[T]) Accept(v Visitor) error {
	return v.Visit(f())
}

func (c *visitOneOf) Visit(v interface{}) error {
	if c.vi == nil {
		return nil
	}
	var n Visitor
	if c.next != nil {
		n = c.next
	}
	return c.vi.visitOrNext(v, n)
}

func (c *visitOneOf) init(f ...OneOf) {
	c.next, c.vi = nil, nil
	it := c
	for i := range f {
		if i == 0 {
			it.vi = f[i]
		} else {
			it.next = &visitOneOf{vi: f[i]}
			it = it.next
		}
	}
}
