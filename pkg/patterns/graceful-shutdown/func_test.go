package graceful_shutdown

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Funcs(t *testing.T) {
	exp := [...]int{1, 2, 3, 4, 5, 6}
	act := [...]int{0, 0, 0, 0, 0, 0}
	st := ForDuration(4*time.Second).Run(
		Func(func(ctx context.Context) {
			act[0] = 1
		}),
		Func1(func(ctx context.Context) error {
			act[1] = 2
			return nil
		}),
		Func2(func(ctx context.Context) (byte, error) {
			act[2] = 3
			return 1, nil
		}),
		Func3(func(ctx context.Context) (byte, int16, error) {
			act[3] = 4
			return 1, 2, nil
		}),
		Func4(func(ctx context.Context) (byte, int16, int32, error) {
			act[4] = 5
			return 1, 2, 3, nil
		}),
		Func5(func(ctx context.Context) (int8, int16, int32, int64, error) {
			act[5] = 6
			return 1, 2, 3, 5, nil
		}),
	)
	require.Equal(t, Completed, st)
	require.Equal(t, exp, act)
}

func Test_FuncAndConsumer(t *testing.T) {
	tt := time.Now()
	e := errors.New("err")

	exp := [...]interface{}{uint(1), "ok", float32(1.2), tt, e}
	act := [...]interface{}{0, 0, 0, 0, 0}

	st := ForDuration(time.Second).Run(
		Func5(func(ctx context.Context) (uint, string, float32, time.Time, error) {
			return uint(1), "ok", float32(1.2), tt, e
		}).Consume(
			func(ctx context.Context, a uint, b string, c float32, d time.Time, e error) {
				act[0] = a
				act[1] = b
				act[2] = c
				act[3] = d
				act[4] = e
			},
		),
	)
	require.Equal(t, Completed, st)
	require.Equal(t, exp, act)
}
