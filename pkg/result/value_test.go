package result

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MatchResultOk(t *testing.T) {
	f := func() (string, error) {
		return "123", nil
	}

	var ret string
	var e error
	Wrap(f()).Ok(func(s string) {
		ret = s
	}).Fail(func(err error) {
		e = err
	})
	require.NoError(t, e)
	require.Equal(t, "123", ret)

	ret, e = "", nil
	Wrap(f()).Fail(func(err error) {
		e = err
	}).Ok(func(s string) {
		ret = s
	})
	require.NoError(t, e)
	require.Equal(t, "123", ret)
}

func Test_MatchResultFail(t *testing.T) {
	f := func() (string, error) {
		return "123", fmt.Errorf("err")
	}

	var ret string
	var e error
	Wrap(f()).Ok(func(s string) {
		ret = s
	}).Fail(func(err error) {
		e = err
	})
	require.Error(t, e)
	require.Equal(t, "", ret)

	ret, e = "", nil
	Wrap(f()).Fail(func(err error) {
		e = err
	}).Ok(func(s string) {
		ret = s
	})
	require.Error(t, e)
	require.Equal(t, "", ret)
}
