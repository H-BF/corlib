package jobs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/patterns/observer"
	"github.com/H-BF/corlib/pkg/patterns/queue"
	"github.com/H-BF/corlib/pkg/scheduler"
	"github.com/H-BF/corlib/pkg/tm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const (
	stageJobSchedulerEnabled = iota + 1
	stageJobSchedulerDisabled
	stageJobSchedulerStarted
	stageJobSchedulerStop
	stageJobSchedulerClose
	stageJobStarted
	stageJobFinished
)

func Test_Case_ErrBackoffStopped(t *testing.T) {
	logger.SetLevel(zap.InfoLevel)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	ctx = tm.TaskManagerToContext(ctx, tm.NewTaskManager())
	defer cancel()

	conf := JobSchedulerConf{
		JobID:         "test-job",
		TaskScheduler: scheduler.NewConstIntervalScheduler(2 * time.Second),
		NewTask: func(_ context.Context) (tm.Task, []interface{}, error) {
			t, e := tm.MakeSimpleTask("1", func() error {
				return errors.New("task is failed")
			})
			return t, nil, e
		},
	}

	finish := make(chan error, 1)
	defer close(finish)
	mx := new(sync.Mutex)
	var gotStageSeq []int

	eventObserve := func(event observer.EventType) {
		switch t := event.(type) {
		case OnJobLog:
			logger.Info(ctx, t)
		case OnJobSchedulerStarted:
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobSchedulerStarted)
			mx.Unlock()
		case OnJobSchedulerEnabled:
			mx.Lock()
			if !t.Enabled {
				gotStageSeq = append(gotStageSeq, stageJobSchedulerDisabled)
			} else {
				gotStageSeq = append(gotStageSeq, stageJobSchedulerEnabled)
			}
			mx.Unlock()
		case OnJobSchedulerStop:
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobSchedulerStop)
			mx.Unlock()
			select {
			case finish <- t.Reason:
			default:
			}
		case OnJobFinished:
			logger.InfoKV(ctx, "event", "scheduler-id", t.JobID, "finished", t)
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobFinished)
			mx.Unlock()
		case OnJobStarted:
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobStarted)
			mx.Unlock()
		}
	}

	schedJob, err := NewJobScheduler(ctx, conf)
	if !assert.NoError(t, err) {
		return
	}
	defer schedJob.Close()
	obs := observer.NewObserver(eventObserve, false)
	SubscribeOnAllEvents(obs)
	schedJob.Subject().ObserversAttach(obs)
	schedJob.Schedule()
	schedJob.Enable(true)

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-finish:
	}

	errors.Is(err, ErrNoTaskIsProvided)
	if !assert.True(t, errors.Is(err, ErrBackoffStopped)) {
		return
	}
	expectedStageSeq := []int{
		stageJobSchedulerStarted,
		stageJobSchedulerEnabled,
		stageJobStarted,
		stageJobFinished,
		stageJobSchedulerStop,
	}
	assert.Equal(t, expectedStageSeq, gotStageSeq)
}

func Test_Case_ErrNoTaskIsProvided(t *testing.T) {
	type task struct {
		tm.Task
		args []interface{}
	}
	mustTask := func(t tm.Task, e error) tm.Task {
		if e != nil {
			panic(e)
		}
		return t
	}
	logger.SetLevel(zap.InfoLevel)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	ctx = tm.TaskManagerToContext(ctx, tm.NewTaskManager())
	defer cancel()
	taskQueue := queue.NewFIFO(ctx)
	ok := taskQueue.Put(
		task{
			Task: mustTask(
				tm.MakeSimpleTask("1", func(a string) string {
					return a
				})),
			args: []interface{}{"TASK1"},
		},
		task{
			Task: mustTask(
				tm.MakeSimpleTask("1", func(a int) int {
					return a
				})),
			args: []interface{}{1},
		},
	)
	if !assert.True(t, ok) {
		return
	}
	_ = taskQueue.Close()

	conf := JobSchedulerConf{
		JobID:         "test-job",
		TaskScheduler: scheduler.NewConstIntervalScheduler(3 * time.Second),
		NewTask: func(_ context.Context) (tm.Task, []interface{}, error) {
			v, err := taskQueue.Get(ctx)
			if err != nil {
				return nil, nil, err
			}
			switch t := v.(type) {
			case task:
				return t.Task, t.args, nil
			default:
				return nil, nil, errors.New("unexpected error")
			}
		},
	}

	finish := make(chan error, 1)
	defer close(finish)
	mx := new(sync.Mutex)
	var gotStageSeq []int

	eventObserve := func(event observer.EventType) {
		switch t := event.(type) {
		case OnJobLog:
			logger.Info(ctx, t)
		case OnJobSchedulerStarted:
			logger.InfoKV(ctx, "scheduler-started", "scheduler-id", t.JobID)
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobSchedulerStarted)
			mx.Unlock()
		case OnJobSchedulerEnabled:
			mx.Lock()
			if !t.Enabled {
				gotStageSeq = append(gotStageSeq, stageJobSchedulerDisabled)
			} else {
				gotStageSeq = append(gotStageSeq, stageJobSchedulerEnabled)
			}
			mx.Unlock()
		case OnJobSchedulerStop:
			logger.InfoKV(ctx, "scheduler-stopped",
				"scheduler-id", t.JobID,
				"reason", t.Reason.Error())
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobSchedulerStop)
			mx.Unlock()
			select {
			case finish <- t.Reason:
			default:
			}
		case OnJobFinished:
			logger.InfoKV(ctx, "event", "scheduler-id", t.JobID, "finished", t)
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobFinished)
			mx.Unlock()
		case OnJobStarted:
			logger.InfoKV(ctx, "event", "scheduler-id", t.JobID, "started", t)
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobStarted)
			mx.Unlock()
		}
	}

	schedJob, err := NewJobScheduler(ctx, conf)
	if !assert.NoError(t, err) {
		return
	}
	defer schedJob.Close()
	obs := observer.NewObserver(eventObserve, false)
	SubscribeOnAllEvents(obs)
	schedJob.Subject().ObserversAttach(obs)
	schedJob.Schedule()
	schedJob.Enable(true)

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-finish:
	}

	errors.Is(err, ErrNoTaskIsProvided)
	if !assert.True(t, errors.Is(err, ErrNoTaskIsProvided)) {
		return
	}

	expectedStageSeq := []int{
		stageJobSchedulerStarted,
		stageJobSchedulerEnabled,
		stageJobStarted,
		stageJobFinished,
		stageJobStarted,
		stageJobFinished,
		stageJobSchedulerStop,
	}
	assert.Equal(t, expectedStageSeq, gotStageSeq)
}

func Test_Case_FullCycle(t *testing.T) {
	var (
		mx          sync.Mutex
		gotStageSeq []int
	)

	logger.SetLevel(zap.InfoLevel)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var (
		schedJob JobScheduler
		err      error
	)
	finish := make(chan error, 1)
	defer close(finish)
	eventObserve := func(event observer.EventType) {
		switch t := event.(type) {
		case OnJobLog:
			logger.Info(ctx, t)
		case OnJobSchedulerStarted:
			logger.InfoKV(ctx, "scheduler-started", "scheduler-id", t.JobID)
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobSchedulerStarted)
			mx.Unlock()
			schedJob.Enable(true)
		case OnJobSchedulerStop:
			logger.InfoKV(ctx, "scheduler-stopped",
				"scheduler-id", t.JobID, "reason", t.Reason.Error())
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobSchedulerStop)
			mx.Unlock()
			select {
			case finish <- t.Reason:
			default:
			}
		case OnJobSchedulerClose:
			logger.InfoKV(ctx, "scheduler-close",
				"scheduler-id", t.JobID)
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobSchedulerClose)
			mx.Unlock()
			select {
			case finish <- nil:
			default:
			}
		case OnJobSchedulerEnabled:
			var what string
			if t.Enabled {
				what = "true"
			} else {
				what = "false"
			}
			logger.InfoKV(ctx, "event", "scheduler-id", t.JobID, "enable", what)
			mx.Lock()
			if !t.Enabled {
				gotStageSeq = append(gotStageSeq, stageJobSchedulerDisabled)
			} else {
				gotStageSeq = append(gotStageSeq, stageJobSchedulerEnabled)
			}
			mx.Unlock()
			if !t.Enabled {
				_ = schedJob.Close()
			}
		case OnJobFinished:
			logger.InfoKV(ctx, "event", "scheduler-id", t.JobID, "finished", t)
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobFinished)
			mx.Unlock()
			schedJob.Enable(false)
		case OnJobStarted:
			logger.InfoKV(ctx, "event", "scheduler-id", t.JobID, "started", t)
			mx.Lock()
			gotStageSeq = append(gotStageSeq, stageJobStarted)
			mx.Unlock()
		}
	}

	ctx = tm.TaskManagerToContext(ctx, tm.NewTaskManager())
	conf := JobSchedulerConf{
		JobID:         "test-job",
		TaskScheduler: scheduler.NewConstIntervalScheduler(10 * time.Second),
		NewTask: func(_ context.Context) (tm.Task, []interface{}, error) {
			tsk, e := tm.MakeSimpleTask("1", func(c context.Context) error {
				select {
				case <-time.After(2 * time.Second):
				case <-c.Done():
					return c.Err()
				}
				return nil
			})
			if e != nil {
				return nil, nil, e
			}
			return tsk, []interface{}{ctx}, nil
		},
	}
	schedJob, err = NewJobScheduler(ctx, conf)
	if !assert.NoError(t, err) {
		return
	}
	defer schedJob.Close()
	obs := observer.NewObserver(eventObserve, false)
	SubscribeOnAllEvents(obs)
	schedJob.Subject().ObserversAttach(obs)
	schedJob.Schedule()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-finish:
	}

	if !assert.NoError(t, err) {
		return
	}

	expectedStageSeq := []int{
		stageJobSchedulerStarted,
		stageJobSchedulerEnabled,
		stageJobStarted,
		stageJobFinished,
		stageJobSchedulerDisabled,
		stageJobSchedulerClose,
	}

	assert.Equal(t, expectedStageSeq, gotStageSeq)
}
