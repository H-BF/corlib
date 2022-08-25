package events

import (
	"runtime"
	"sync"
)

//EventID id of event
type EventID int

//Event ...
type Event interface {
	ID() EventID
	Fire()
	FireIf(val bool)
	HasFired() bool
	Done() <-chan struct{}
}

//NewEvent ...
func NewEvent(id EventID) Event {
	var once sync.Once
	ch := make(chan struct{})
	ret := &eventImpl{
		id:     id,
		marker: ch,
		fireOnce: func() {
			once.Do(func() {
				close(ch)
			})
		},
	}
	runtime.SetFinalizer(ret, func(o *eventImpl) {
		o.fireOnce()
	})
	return ret
}

var (
	_ Event = (*eventImpl)(nil)
	_       = NewEvent
)

type eventImpl struct {
	id       EventID
	marker   chan struct{}
	fireOnce func()
}

//ID ...
func (ev *eventImpl) ID() EventID {
	return ev.id
}

//Fire ...
func (ev *eventImpl) Fire() {
	ev.fireOnce()
}

func (ev *eventImpl) FireIf(val bool) {
	if val {
		ev.fireOnce()
	}
}

//HasFired ...
func (ev *eventImpl) HasFired() bool {
	select {
	case <-ev.marker:
		return true
	default:
	}
	return false
}

func (ev *eventImpl) Done() <-chan struct{} {
	return ev.marker
}
