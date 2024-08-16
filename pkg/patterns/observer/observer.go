package observer

import (
	"reflect"
	"sync"
)

// NewObserver создает обозреватель событий
func NewObserver(er EventReceiver, async bool, events ...EventType) Observer {
	ret := &observerImpl{
		EventReceiver: er,
		async:         async,
		regEvents:     make(map[reflect.Type]struct{}),
	}
	ret.SubscribeEvents(events...)
	return ret
}

type observerImpl struct {
	mx sync.RWMutex
	EventReceiver
	closed    bool
	async     bool
	regEvents map[reflect.Type]struct{}
}

func (o *observerImpl) Close() error {
	o.mx.Lock()
	defer o.mx.Unlock()
	if !o.closed {
		o.closed = true
		o.EventReceiver = nil
		o.regEvents = nil
	}
	return nil
}

func (o *observerImpl) SubscribeEvents(events ...EventType) {
	o.mx.Lock()
	defer o.mx.Unlock()
	if o.closed {
		return
	}
	for i := range events {
		o.regEvents[reflect.TypeOf(events[i])] = struct{}{}
	}
}

func (o *observerImpl) UnsubscribeEvents(events ...EventType) {
	o.mx.Lock()
	defer o.mx.Unlock()
	if o.closed {
		return
	}
	for i := range events {
		delete(o.regEvents, reflect.TypeOf(events[i]))
	}
}

func (o *observerImpl) UnsubscribeAllEvents() {
	o.mx.Lock()
	defer o.mx.Unlock()
	if o.closed {
		return
	}
	o.regEvents = make(map[reflect.Type]struct{})
}

// Observe impl observer
func (o *observerImpl) Observe(events ...EventType) {
	if len(events) == 0 {
		return
	}
	var filteredEvents []EventType
	o.mx.RLock()
	if !o.closed {
		filteredEvents = make([]EventType, 0, len(events))
		for _, e := range events {
			if _, can := o.regEvents[reflect.TypeOf(e)]; can {
				filteredEvents = append(filteredEvents, e)
			}
		}
	}
	o.mx.RUnlock()
	for i := range filteredEvents {
		if o.async {
			i := i
			go o.EventReceiver(filteredEvents[i])
		} else {
			o.EventReceiver(filteredEvents[i])
		}
	}
}
