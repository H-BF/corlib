package observer

// NullObserver does nothing
type NullObserver struct{}

var _ Observer = (*NullObserver)(nil)

// Close impl Observer iface
func (NullObserver) Close() error {
	return nil
}

// SubscribeEvents impl Observer iface
func (NullObserver) SubscribeEvents(...EventType) {}

// UnsubscribeEvents impl Observer iface
func (NullObserver) UnsubscribeEvents(...EventType) {}

// UnsubscribeAllEvents impl Observer iface
func (NullObserver) UnsubscribeAllEvents() {}

// Observe impl Observer iface
func (NullObserver) Observe(...EventType) {}
