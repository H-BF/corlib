package grpc

import (
	"context"
	"io"
	"testing"

	"github.com/H-BF/corlib/client/grpc/internal"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestWithErrorWrapper(t *testing.T) {
	pure := new(internal.InvalidConn)
	c := WithErrorWrapper(pure, "")
	_, ok := c.(errWrapperInterface)
	if !assert.True(t, ok) {
		return
	}

	var closerClient *grpc.ClientConn
	c = WithErrorWrapper(closerClient, "")
	_, ok = c.(io.Closer)
	if !assert.True(t, ok) {
		return
	}
	_, ok = c.(errWrapperInterface)
	if !assert.True(t, ok) {
		return
	}

	closable := MakeCloseable(closerClient)
	c = WithErrorWrapper(closable, "")
	_, ok = c.(Closable)
	if !assert.True(t, ok) {
		return
	}
	_, ok = c.(errWrapperInterface)
	if !assert.True(t, ok) {
		return
	}

	c1 := WithErrorWrapper(MakeCloseable(pure), "")
	e := c1.(Closable).CloseConn()
	if !assert.NoError(t, e) {
		return
	}
	ctx := context.Background()
	e = c1.Invoke(ctx, "/service1/method1", nil, nil)
	assert.ErrorIs(t, e, ErrConnClosed)
}
