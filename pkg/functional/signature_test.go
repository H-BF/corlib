package functional

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type forTestSignature struct {
	n int
}

func (ft *forTestSignature) foo()             {}
func (ft *forTestSignature) foo1(int)         {}
func (ft *forTestSignature) foo2(int, string) {}
func (ft *forTestSignature) foo3(...int)      {}
func (ft *forTestSignature) foo4(int, ...string) {
	ft.n++
}

func Test_Signature(t *testing.T) {
	type simple struct {
		fun      interface{}
		variadic bool
		nArgs    int
	}
	obj := forTestSignature{}
	samples := []simple{
		{fun: func() {}, variadic: false, nArgs: 0},
		{fun: func(int) {}, variadic: false, nArgs: 1},
		{fun: func(int, string) {}, variadic: false, nArgs: 2},
		{fun: func(...int) {}, variadic: true, nArgs: 1},
		{fun: func(int, ...string) {}, variadic: true, nArgs: 2},
		{fun: obj.foo, variadic: false, nArgs: 0},
		{fun: obj.foo1, variadic: false, nArgs: 1},
		{fun: obj.foo2, variadic: false, nArgs: 2},
		{fun: obj.foo3, variadic: true, nArgs: 1},
		{fun: obj.foo4, variadic: true, nArgs: 2},
	}
	for i := range samples {
		sample := samples[i]
		S, err := MaySignatureOf(sample.fun)
		assert.NoError(t, err)
		args, variadic := S.ArgsInfo()
		assert.Equal(t, sample.nArgs, len(args), "arg-count")
		assert.Equal(t, sample.variadic, variadic, "variadic")
	}

	S1 := MustSignatureOf(obj.foo4)
	S2 := MustSignatureOf(func(int, ...string) {})
	equal := S1.EqualTo(S2)
	assert.Equal(t, true, equal, "equal-signatures")
}
