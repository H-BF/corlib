package observer

// NewTiedSubj -
func NewTiedSubj(tied Subject) Subject {
	return &tiedSubject{
		Subject: NewSubject(),
		tied:    tied,
	}
}

type tiedSubject struct {
	Subject
	tied Subject
}

var _ Subject = (*tiedSubject)(nil)

// Notify impl observer.Subject iface
func (sb *tiedSubject) Notify(events ...EventType) {
	sb.Subject.Notify(events...)
	sb.tied.Notify(events...)
}
