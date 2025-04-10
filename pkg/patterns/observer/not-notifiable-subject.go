package observer

// NotNotifiableSubject has empty Notify method
type NotNotifiableSubject struct {
	Subject
}

// Notify impl Subject iface
func (NotNotifiableSubject) Notify(...EventType) {}
