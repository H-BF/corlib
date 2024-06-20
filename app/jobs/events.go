package jobs

import (
	"errors"
	"fmt"
	"time"

	"github.com/H-BF/corlib/pkg/patterns/observer"
)

var ( //errors
	//ErrNoTaskManager no task manager
	ErrNoTaskManager = errors.New("no any 'TaskManager' is provided")

	//ErrNoTaskIsProvided task is not created
	ErrNoTaskIsProvided = errors.New("no 'Task' is provided any more")

	//ErrBackoffStopped when backoff stops its evaluating
	ErrBackoffStopped = errors.New("backoff is stopped")
)

// SubscribeOnAllEvents subscribe on all events
func SubscribeOnAllEvents(obs observer.Observer) {
	obs.SubscribeEvents([]observer.EventType{
		OnJobSchedulerClose{},
		OnJobSchedulerStarted{},
		OnJobSchedulerEnabled{},
		OnJobSchedulerStop{},
		OnJobStarted{},
		OnJobFinished{},
		OnJobLog{},
	}...)
}

// OnJobLog info / debug log
type OnJobLog struct {
	observer.EventType `json:"-"`
	observer.TextMessageEvent
	JobID string
}

func (evt OnJobLog) String() string {
	return fmt.Sprintf("[%s]: %s", evt.JobID, evt.TextMessageEvent)
}

// OnJobSchedulerEnabled when job enabled
type OnJobSchedulerEnabled struct {
	observer.EventType `json:"-"`
	JobID              string
	Enabled            bool
}

// OnJobSchedulerStarted when scheduler activates
type OnJobSchedulerStarted struct {
	observer.EventType `json:"-"`
	JobID              string
}

// OnJobStarted subject event from PeriodicJobScheduler
type OnJobStarted struct {
	observer.EventType `json:"-"`
	JobID              string
	At                 time.Time //UTC
}

// OnJobFinished subject event from PeriodicJobScheduler
type OnJobFinished struct {
	observer.EventType `json:"-"`
	JobResult
	JobID string
	At    time.Time //UTC

	errorIndex int
}

func (evt *OnJobFinished) setResult(jobOutput []interface{}, jobStartFailure error) {
	if jobStartFailure != nil {
		evt.JobResult = JobStartFailure{StartFailure: errMarshal{error: jobStartFailure}}
	} else {
		evt.errorIndex = len(jobOutput)
		for i := 0; i < evt.errorIndex; i++ {
			if _, ok := jobOutput[i].(error); ok {
				evt.errorIndex = i
				break
			}
		}
		if evt.errorIndex < len(jobOutput) {
			var o JobOutput
			o.Output = append(o.Output, jobOutput...)
			o.Output[evt.errorIndex] = errMarshal{error: o.Output[evt.errorIndex].(error)}
			evt.JobResult = o
		} else {
			evt.JobResult = JobOutput{Output: jobOutput}
		}
	}
}

// FindError ,..,
func (evt OnJobFinished) FindError() error {
	switch t := evt.JobResult.(type) {
	case JobStartFailure:
		return t.StartFailure
	case JobOutput:
		if evt.errorIndex < len(t.Output) {
			return t.Output[evt.errorIndex].(error)
		}
	}
	return nil
}

// OnJobSchedulerClose when scheduler is about to be closed from Close method
type OnJobSchedulerClose struct {
	observer.EventType `json:"-"`
	JobID              string
}

// OnJobSchedulerStop when jeb scheduler is about to be stopped
type OnJobSchedulerStop struct {
	observer.EventType `json:"-"`
	JobID              string
	Reason             error
}

// JobResult ...
type JobResult interface {
	isJobResult()
}

// JobStartFailure ...
type JobStartFailure struct {
	JobResult    `json:"-"`
	StartFailure error `json:",omitempty"` //nolint:tagliatelle
}

// JobOutput ...
type JobOutput struct {
	JobResult `json:"-"`
	Output    []interface{} `json:",omitempty"` //nolint:tagliatelle
}
