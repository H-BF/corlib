package graceful_shutdown

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Chan(t *testing.T) {
	tt := time.Now()
	var tt2 time.Time
	ch := make(chan time.Time)
	go func() {
		defer close(ch)
		select {
		case <-time.NewTimer(time.Second).C:
		case ch <- tt:
		}
	}()
	st := ForDuration(3 * time.Second).Run(
		Chan(ch).Consume(func(ctx context.Context, a time.Time) {
			tt2 = a
		}),
	)
	require.Equal(t, Completed, st)
	require.Equal(t, tt, tt2)

	tt = time.Now().Add(time.Second)
	st = ForDuration(time.Second).Run(
		Chan(ch).Consume(func(ctx context.Context, a time.Time) {
			tt2 = a
		}).IfNoResult(func() time.Time {
			return tt
		}),
	)
	require.Equal(t, Completed, st)
	require.Equal(t, tt, tt2)
}
