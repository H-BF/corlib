package observer

// ComposeWithObservees -
func ComposeWithObservees(s Subject, observees ...Observee) Subject {
	if len(observees) == 0 {
		return s
	}
	return &subjectAndObserveesComopsit{
		Subject:   s,
		observees: observees,
	}
}

type subjectAndObserveesComopsit struct {
	Subject
	observees []Observee
}

var _ Subject = (*subjectAndObserveesComopsit)(nil)

// Notify impl observer.Subject iface
func (sb *subjectAndObserveesComopsit) Notify(events ...EventType) {
	sb.Subject.Notify(events...)
	for i := range sb.observees {
		sb.observees[i].Notify(events...)
	}
}
