package observer

// NullSubject it does nothing
type NullSubject struct{}

var _ Subject = (*NullSubject)(nil)

// ObserversAttach impl Subject iface
func (NullSubject) ObserversAttach(...Observer) {}

// ObserversAttach  impl Subject iface
func (NullSubject) ObserversDetach(...Observer) {}

// ObserversAttach  impl Subject iface
func (NullSubject) DetachAllObservers() {}

// ObserversAttach  impl Subject iface
func (NullSubject) Notify(...EventType) {}

// Close impl Subject iface
func (NullSubject) Close() error { return nil }
