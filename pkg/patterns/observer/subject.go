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
		closed bool
	}
)

// ObserversAttach impl Subject iface
func (s *subjectImpl) ObserversAttach(observers ...Observer) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if !s.closed {
		for _, o := range observers {
			s.observerHolder[o] = struct{}{}
		}
	}
}

// ObserversDetach impl Subject iface
func (s *subjectImpl) ObserversDetach(observers ...Observer) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if !s.closed {
		for _, o := range observers {
			delete(s.observerHolder, o)
		}
	}
}

// DetachAllObservers impl Subject iface
func (s *subjectImpl) DetachAllObservers() {
	s.mx.Lock()
	if !s.closed {
		s.observerHolder = make(observerHolder)
	}
	s.mx.Unlock()
}

// Notify impl Subject iface
func (s *subjectImpl) Notify(events ...EventType) {
	if len(events) > 0 {
		observers := s.observerList()
		for i := range observers {
			observers[i].Observe(events...)
		}
	}
}

// Close impl Subject iface
func (s *subjectImpl) Close() error {
	s.mx.Lock()
	s.closed, s.observerHolder = true, nil
	s.mx.Unlock()
	return nil
}

func (s *subjectImpl) observerList() []Observer {
	s.mx.RLock()
	defer s.mx.RUnlock()
	observers := make([]Observer, 0, len(s.observerHolder))
	for o := range s.observerHolder {
		observers = append(observers, o)
	}
	return observers
}
