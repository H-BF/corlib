package visitor

type (
	// Acceptor it is visitor` acceptor def
	Acceptor interface {
		Accept(Visitor) error
	}

	// Visitor it is visitor def
	Visitor interface {
		Visit(interface{}) error
	}

	// OneOf base interface
	OneOf interface {
		visitOrNext(v interface{}, next Visitor) error
	}
)

// Visit it tries to visit accessor by all visitors
func Visit(acc Acceptor, vis ...Visitor) (err error) {
	for _, v := range vis {
		if err = acc.Accept(v); err != nil {
			break
		}
	}
	return err
}

// Func2Visitor it makes function as Visitor && VisitorOneOf
func Func2Visitor[T any](f func(T) error) visitorFunc[T] {
	return f
}

// NewOneOf it makes one-of visitor
func NewOneOf(vs ...OneOf) Visitor {
	ret := new(visitOneOf)
	ret.init(vs...)
	return ret
}
