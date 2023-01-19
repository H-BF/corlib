package graceful_shutdown

import (
	"context"
	"reflect"
	"time"
)

// ForDuration makes runner with max time duration
func ForDuration(d time.Duration) runner { //nolint:revive
	return runner{maxDuration: d}
}

//                   --== IMPL ==--

type runner struct {
	maxDuration time.Duration
	attrs       []contextAttr
}

// WithContextAttrs add some attrs to context
func (run runner) WithContextAttrs(attrs ...contextAttr) runner {
	ret := run
	ret.attrs = append(ret.attrs, attrs...)
	return ret
}

// Run it runs task(s)
func (run runner) Run(tasks ...Task) Status {
	if run.maxDuration <= 0 {
		return Timeout
	}
	if len(tasks) == 0 {
		return Completed
	}
	ctx := context.Background()
	for i := range run.attrs {
		ctx = run.attrs[i].imbue(ctx)
	}
	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, run.maxDuration)
	defer cancel()

	cases := append(make([]reflect.SelectCase, 0, len(tasks)+1),
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		},
	)
	for _, it := range tasks {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(it.Exec(ctx)),
		})
	}
	for len(cases) > 1 {
		selected, _, _ := reflect.Select(cases)
		if selected == 0 { //maxDuration has gone
			return Timeout
		}
		cases = append(
			append(cases[:0], cases[:selected]...),
			cases[selected+1:]...,
		)
	}
	return Completed
}
