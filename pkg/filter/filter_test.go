package filter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type inacceptableScope struct {
	Scope
}

func (inacceptableScope) intt(int) bool {
	return true
}

func (inacceptableScope) Uintt(uint) bool {
	return true
}

func (inacceptableScope) Uintt64(int64) bool {
	return true
}

type statefulScope struct {
	Scope
	int
}

func (s statefulScope) Pass(v int) bool {
	return s.int == v
}

func TestLogics(t *testing.T) {
	type tcase struct {
		sc   Scope
		fail bool
		pass bool
	}

	cases := [...]tcase{
		{inacceptableScope{}, true, false},
		{Not(inacceptableScope{}), true, false},

		{NoScope{}, false, true},

		{Not(NoScope{}), false, false},
		{Not(Not(NoScope{})), false, true},

		{And(NoScope{}, NoScope{}), false, true},
		{And(Not(NoScope{}), NoScope{}), false, false},
		{And(NoScope{}, Not(NoScope{})), false, false},
		{And(Not(NoScope{}), Not(NoScope{})), false, false},

		{Or(NoScope{}, NoScope{}), false, true},
		{Or(Not(NoScope{}), NoScope{}), false, true},
		{Or(NoScope{}, Not(NoScope{})), false, true},
		{Or(Not(NoScope{}), Not(NoScope{})), false, false},

		{ScopeFromFunc(func(int) bool { return true }), false, true},
		{ScopeFromFunc(func(int) bool { return false }), false, false},

		{All(), false, false},
		{All(NoScope{}, NoScope{}), false, true},
		{All(NoScope{}, Not(NoScope{})), false, false},
		{All(Not(NoScope{}), NoScope{}), false, false},

		{Any(), false, false},
		{Any(Not(NoScope{}), NoScope{}), false, true},
		{Any(NoScope{}, Not(NoScope{})), false, true},
		{Any(Not(NoScope{}), Not(NoScope{})), false, false},

		{statefulScope{}, false, false},
		{statefulScope{int: 1}, false, true},
	}
	for i := range cases {
		msg := fmt.Sprintf("on test case #%v", i)
		c := cases[i]
		var f SimpleFilter[int]
		e := f.InitFromScope(c.sc)
		if c.fail {
			require.Error(t, e, msg)
		} else {
			require.NoError(t, e, msg)
			pass := f(1)
			require.Equal(t, c.pass, pass, msg)
		}
	}
}
