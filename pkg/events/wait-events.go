package events

import (
	"context"
	"reflect"
)

// WaitOne ждем когда сработает хотябы одно слбытие
func WaitOne(ctx context.Context, event1 Event, otherEvents ...Event) (int, error) {
	n := 1 + len(otherEvents)
	selectCases := make([]reflect.SelectCase, 0, n+1)
	for _, ev := range append([]Event{event1}, otherEvents...) {
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ev.Done()),
		})
	}
	selectCases = append(selectCases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	})
	selected, _, _ := reflect.Select(selectCases)
	if selected == n {
		return selected, ctx.Err()
	}
	return selected, nil
}

// WaitAll ждем когда сработают все события
func WaitAll(ctx context.Context, event1 Event, otherEvents ...Event) error {
	n := 1 + len(otherEvents)
	selectCases := make([]reflect.SelectCase, 0, n+1)
	for _, ev := range append([]Event{event1}, otherEvents...) {
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ev.Done()),
		})
	}
	selectCases = append(selectCases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	})

	var (
		count    int
		ch       chan struct{}
		selected int
	)
	for count < n {
		if selected, _, _ = reflect.Select(selectCases); selected == n {
			return ctx.Err()
		}
		count++
		selectCases[selected].Chan = reflect.ValueOf(ch)
	}
	return nil
}

var (
	_ = WaitOne
	_ = WaitAll
)
