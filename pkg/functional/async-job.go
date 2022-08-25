package functional

import (
	"context"

	"github.com/pkg/errors"
)

//AsyncJobStatus async job status
type AsyncJobStatus uint8

const (
	//AsyncJobRunning job is running for now
	AsyncJobRunning AsyncJobStatus = iota + 1
	//AsyncJobCompleted job has completed
	AsyncJobCompleted
	//AsyncJobStartFailed job has failed at start
	AsyncJobStartFailed
)

//AsyncJobState job state at moment
type AsyncJobState struct {
	Status   AsyncJobStatus
	Returned []interface{}
	Failure  error
}

//AsyncJobControl job control interface
type AsyncJobControl interface {
	Wait(ctx context.Context) (AsyncJobState, error)
}

//AsyncJob asynchronous job interface
type AsyncJob interface {
	Run(args ...interface{}) AsyncJobControl
	WhenStateChanged(catchers ...func(state AsyncJobState)) AsyncJob
	WhenCompleted(funcResultReceivers ...interface{}) AsyncJob
}

//MustAsyncJob construct async job or panic if error
func MustAsyncJob(f interface{}) AsyncJob {
	job, e := MayAsyncJob(f)
	if e != nil {
		panic(e)
	}
	return job
}

//MayAsyncJob construct async job or return error
func MayAsyncJob(f interface{}) (AsyncJob, error) {
	const api = "MayAsyncJob"
	if j, ok := f.(AsyncJob); ok {
		return j, nil
	}
	c, e := MayCallableOf(f)
	if e != nil {
		return nil, errors.Wrap(e, api)
	}
	return &asyncJob{job: c}, nil
}

type asyncJobControl struct {
	barrier      chan struct{}
	commonResult interface{} // --> []interface{} | error
}

//Wait wait result from async job
func (jc *asyncJobControl) Wait(ctx context.Context) (AsyncJobState, error) {
	const api = "AsyncJobControl/Wait"
	var result AsyncJobState
	var err error
	select {
	case <-jc.barrier:
		switch t := jc.commonResult.(type) {
		case error:
			result.Status = AsyncJobStartFailed
			result.Failure = t
		case []interface{}:
			result.Status = AsyncJobCompleted
			result.Returned = t
		default:
			panic(errors.Errorf("%s: unexpected behaviour reached", api))
		}
	case <-ctx.Done():
		err = ctx.Err()
	}
	return result, err
}

type asyncJob struct {
	job                Callable
	jobResultReceivers []Callable
	jobStateReceivers  []func(sate AsyncJobState)
}

//WhenCompleted register job result receiver callbacks
func (job *asyncJob) WhenCompleted(funcResultReceivers ...interface{}) AsyncJob {
	const api = "asyncJob/WhenCompleted"
	expectedSignature := job.job.FromOutputValues()
	var recv []Callable
	for _, f := range funcResultReceivers {
		c, e := MayCallableOf(f)
		if e != nil {
			panic(errors.Wrap(e, api))
		}
		if !expectedSignature.EqualTo(c) {
			panic(errors.Errorf("%s: incorrect signature for job result receiver", api))
		}
		recv = append(recv, c)
	}
	cpy := *job
	cpy.jobResultReceivers = recv
	return &cpy
}

//WhenStateChanged register callbacks to receive changes job state
func (job *asyncJob) WhenStateChanged(catchers ...func(sate AsyncJobState)) AsyncJob {
	cpy := *job
	cpy.jobStateReceivers = catchers
	return &cpy
}

//Run start async job
func (job *asyncJob) Run(args ...interface{}) AsyncJobControl {
	const api = "asyncJob/Run"
	control := &asyncJobControl{
		barrier: make(chan struct{}),
	}
	setFinalState := func(vals []interface{}, e error) {
		st := AsyncJobState{
			Returned: vals,
			Failure:  e,
		}
		if e != nil {
			st.Status = AsyncJobStartFailed
			control.commonResult = e
		} else {
			st.Status = AsyncJobCompleted
			control.commonResult = vals
		}
		close(control.barrier)
		for _, f := range job.jobStateReceivers {
			f(st)
		}
	}
	go func() {
		res, e := job.job.Invoke(args...)
		setFinalState(res, errors.Wrap(e, api))
		if e == nil {
			for _, f := range job.jobResultReceivers {
				MustInvokeNoResult(f, res...)
			}
		}
	}()
	return control
}
