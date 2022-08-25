package functional

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type forTestCallable struct {
	data int
}

func (tc *forTestCallable) add(n int) int {
	tc.data += n
	return tc.data
}

func Test_Callable(t *testing.T) {
	var err error
	var called bool
	f0 := func() {
		called = true
	}

	callable := MustCallableOf(f0)
	_, err = callable.Invoke()
	assert.NoError(t, err, "call-no-arg")
	assert.Equal(t, true, called, "call-no-arg")

	f1 := func(n int) {
		called = true
	}
	called = false
	callable = MustCallableOf(f1)
	_, err = callable.Invoke(1)
	assert.NoError(t, err, "call-1-arg")
	assert.Equal(t, true, called, "call-1-arg")

	f2 := func(n int, m map[string]int) {
		called = true
	}
	called = false
	callable = MustCallableOf(f2)
	_, err = callable.Invoke(1, nil)
	assert.NoError(t, err, "call-2-arg")
	assert.Equal(t, true, called, "call-2-arg")

	called = false
	_, err = callable.Invoke(1, map[string]int{})
	assert.NoError(t, err, "call-2-arg")
	assert.Equal(t, true, called, "call-2-arg")

	countArgs := 0
	called = false
	f3 := func(args ...int) {
		countArgs = len(args)
		called = true
	}
	callable = MustCallableOf(f3)
	_, err = callable.Invoke()
	assert.NoError(t, err, "call-no-arg-variadic")
	assert.Equal(t, true, called, "call-no-arg-variadic")
	assert.Equal(t, 0, countArgs, "call-no-arg-variadic")

	countArgs = 0
	called = false
	_, err = callable.Invoke(1)
	assert.NoError(t, err, "call-no-arg-variadic")
	assert.Equal(t, true, called, "call-no-arg-variadic")
	assert.Equal(t, 1, countArgs, "call-no-arg-variadic")

	f4 := func(s string, args ...int) {
		countArgs = len(args) + 1
		called = true
	}
	callable = MustCallableOf(f4)
	countArgs = 0
	called = false
	_, err = callable.Invoke("")
	assert.NoError(t, err, "call-1-arg-variadic")
	assert.Equal(t, true, called, "call-1-arg-variadic")
	assert.Equal(t, 1, countArgs, "call-1-arg-variadic")

	countArgs = 0
	called = false
	_, err = callable.Invoke("", 1)
	assert.NoError(t, err, "call-1-arg-variadic")
	assert.Equal(t, true, called, "call-1-arg-variadic")
	assert.Equal(t, 2, countArgs, "call-1-arg-variadic")

	var d forTestCallable
	callable = MustCallableOf(d.add)
	_, err = callable.Invoke(100)
	assert.NoError(t, err, "method-call-1-arg")
	assert.Equal(t, 100, d.data, "call-1-arg-variadic")
}
