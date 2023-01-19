package graceful_shutdown

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Attrs(t *testing.T) {
	type (
		k1 struct{}
		k2 struct{}
		k3 struct{}
	)
	tt := time.Now()
	exp := [...]interface{}{"ok", 10, tt}
	act := [...]interface{}{0, 0, 0}
	st := ForDuration(time.Second).
		WithContextAttrs(
			Attr(k1{}, "ok"),
			Attr(k2{}, 10),
			Attr(k3{}, tt),
		).
		Run(Func(func(ctx context.Context) {
			act[0] = ctx.Value(k1{})
			act[1] = ctx.Value(k2{})
			act[2] = ctx.Value(k3{})
		}))
	require.Equal(t, Completed, st)
	require.Equal(t, exp, act)
}
