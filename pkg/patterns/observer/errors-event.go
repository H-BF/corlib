package observer

import (
	"go.uber.org/multierr"
)

// ErrorsEvent errors event
type ErrorsEvent struct {
	EventType
	Errors []error
}

// AddErrors add an error
func (ee ErrorsEvent) AddErrors(errs ...error) ErrorsEvent {
	return ErrorsEvent{
		Errors: append(ee.Errors, errs...),
	}
}

// Combined make combined error from errors
func (ee ErrorsEvent) Combined() error {
	return multierr.Combine(ee.Errors...)
}
