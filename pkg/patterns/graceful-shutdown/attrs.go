package graceful_shutdown

import (
	"context"
)

type (
	contextAttr interface {
		imbue(context.Context) context.Context
	}

	contextAttrImpl[K comparable, T interface{}] struct {
		key K
		val T
	}
)

// Attr makes context attribute
func Attr[K comparable, T interface{}](key K, val T) contextAttr {
	return contextAttrImpl[K, T]{key: key, val: val}
}

func (attr contextAttrImpl[K, T]) imbue(ctx context.Context) context.Context {
	return context.WithValue(ctx, attr.key, attr.val)
}
