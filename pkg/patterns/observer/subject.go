package observer

import (
	"sync"
)

// NewSubject создает субъект для оповещения обозревателей событий
func NewSubject() Subject {
	return &subjectImpl{
		observerHolder: make(observerHolder),
	}
}

type (
	observerHolder map[Observer]struct{}
	subjectImpl    struct {
		mx sync.RWMutex
		observerHolder
	}
)

func (s *subjectImpl) ObserversAttach(observers ...Observer) {
	s.mx.Lock()
	defer s.mx.Unlock()
	for _, o := range observers {
		s.observerHolder[o] = struct{}{}
	}
}

func (s *subjectImpl) ObserversDetach(observers ...Observer) {
	s.mx.Lock()
	defer s.mx.Unlock()
	for _, o := range observers {
		delete(s.observerHolder, o)
	}
}

func (s *subjectImpl) DetachAllObservers() {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.observerHolder = make(observerHolder)
}

func (s *subjectImpl) Notify(events ...EventType) {
	if len(events) > 0 {
		s.mx.RLock()
		observers := make([]Observer, 0, len(s.observerHolder))
		for o := range s.observerHolder {
			observers = append(observers, o)
		}
		s.mx.RUnlock()
		for i := range observers {
			observers[i].Observe(events...)
		}
	}
}
