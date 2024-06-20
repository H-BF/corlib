package observer

import (
	"reflect"
	"sync"
)

type (
	//EventType тип сообщения
	EventType interface {
		isObserverEventType()
	}
	//EventReceiver получалель сообщений
	EventReceiver func(event EventType)
	//Observer тот кто получит сообщения
	Observer interface {
		Close() error
		SubscribeEvents(...EventType)
		UnsubscribeEvents(...EventType)
		UnsubscribeAllEvents()
		observe(...EventType)
		canRunAsync() bool
	}
	//Subject источник сообщений
	Subject interface {
		ObserversAttach(...Observer)
		ObserversDetach(...Observer)
		DetachAllObservers()
		Notify(...EventType)
	}
)

// NewSubject создает субъект для оповещения обозревателей событий
func NewSubject() Subject {
	return &subjectImpl{
		observerHolder: make(observerHolder),
	}
}

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

// ------------------------------I M P L--------------------------------

type (
	observerHolder map[Observer]struct{}
	subjectImpl    struct {
		sync.RWMutex
		observerHolder
	}
	observerImpl struct {
		sync.RWMutex
		EventReceiver
		closed    bool
		async     bool
		regEvents map[reflect.Type]struct{}
	}
)

func (s *subjectImpl) ObserversAttach(observers ...Observer) {
	s.Lock()
	defer s.Unlock()
	for _, o := range observers {
		s.observerHolder[o] = struct{}{}
	}
}

func (s *subjectImpl) ObserversDetach(observers ...Observer) {
	s.Lock()
	defer s.Unlock()
	for _, o := range observers {
		delete(s.observerHolder, o)
	}
}

func (s *subjectImpl) DetachAllObservers() {
	s.Lock()
	defer s.Unlock()
	s.observerHolder = make(observerHolder)
}

func (s *subjectImpl) Notify(events ...EventType) {
	if len(events) == 0 {
		return
	}
	broadcast := func(obs []Observer) {
		for i := range obs {
			obs[i].observe(events...)
		}
	}
	s.RLock()
	observers := make([]Observer, 0, len(s.observerHolder))
	async := make([]Observer, 0, len(s.observerHolder))
	for o := range s.observerHolder {
		if o.canRunAsync() {
			async = append(async, o)
		} else {
			observers = append(observers, o)
		}
	}
	s.RUnlock()
	if len(async) > 0 {
		go broadcast(async)
	}
	broadcast(observers)
}

// --------------------------------------============  Observer IMPL ============--------------------------------------

func (o *observerImpl) Close() error {
	o.Lock()
	defer o.Unlock()
	if !o.closed {
		o.closed = true
		o.EventReceiver = nil
		o.regEvents = nil
	}
	return nil
}

func (o *observerImpl) SubscribeEvents(events ...EventType) {
	o.Lock()
	defer o.Unlock()
	if o.closed {
		return
	}
	for i := range events {
		o.regEvents[reflect.TypeOf(events[i])] = struct{}{}
	}
}

func (o *observerImpl) UnsubscribeEvents(events ...EventType) {
	o.Lock()
	defer o.Unlock()
	if o.closed {
		return
	}
	for i := range events {
		delete(o.regEvents, reflect.TypeOf(events[i]))
	}
}

func (o *observerImpl) UnsubscribeAllEvents() {
	o.Lock()
	defer o.Unlock()
	if o.closed {
		return
	}
	o.regEvents = make(map[reflect.Type]struct{})
}

func (o *observerImpl) observe(events ...EventType) {
	if len(events) == 0 {
		return
	}
	var filteredEvents []EventType
	o.RLock()
	if !o.closed {
		filteredEvents = make([]EventType, 0, len(events))
		for _, e := range events {
			if _, can := o.regEvents[reflect.TypeOf(e)]; can {
				filteredEvents = append(filteredEvents, e)
			}
		}
	}
	o.RUnlock()
	for i := range filteredEvents {
		o.EventReceiver(filteredEvents[i])
	}
}

func (o *observerImpl) canRunAsync() bool {
	return o.async
}
