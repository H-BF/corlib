package visitor

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestVisitor(t *testing.T) {
	type (
		tA struct{}
		tB struct{}
		tC struct{}
	)

	var exp interface{}

	vi := NewOneOf(
		Func2Visitor(func(o tA) error {
			exp = o
			return nil
		}),
		Func2Visitor(func(o tB) error {
			exp = o
			return nil
		}),
		Func2Visitor(func(o tC) error {
			exp = o
			return nil
		}),
	)

	require.NoError(t, Visit(WrapAsAcceptor(tA{}), vi))
	require.Equal(t, exp, tA{})

	require.NoError(t, Visit(WrapAsAcceptor(tB{}), vi))
	require.Equal(t, exp, tB{})

	require.NoError(t, Visit(WrapAsAcceptor(tC{}), vi))
	require.Equal(t, exp, tC{})

	customErr := errors.New("err1")
	vi = NewOneOf(
		Func2Visitor(func(tA) error {
			return customErr
		}),
	)
	require.ErrorIs(t, Visit(WrapAsAcceptor(tA{}), vi), customErr)

	tt := time.Now()
	var exp2 interface{}
	err := Visit(WrapAsAcceptor(tt),
		Func2Visitor(func(o time.Time) error {
			exp = o
			return nil
		}),
		Func2Visitor(func(o fmt.Stringer) error {
			exp2 = o
			return nil
		}),
	)
	require.NoError(t, err)
	require.Equal(t, exp, tt)
	require.Equal(t, exp2, interface{}(tt).(fmt.Stringer))
}
