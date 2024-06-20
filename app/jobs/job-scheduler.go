package jobs

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/H-BF/corlib/pkg/backoff"
	"github.com/H-BF/corlib/pkg/patterns/observer"
	"github.com/H-BF/corlib/pkg/patterns/queue"
	"github.com/H-BF/corlib/pkg/scheduler"
	"github.com/H-BF/corlib/pkg/tm"
	"github.com/pkg/errors"
)

const (
	minDelayBeforeStart = time.Second
	logTimeLayout       = "2006-01-02T15:04:05.9Z07"
)

type (
	//JobSchedulerConf провайдер периодических задач
	JobSchedulerConf struct {
		JobID         string
		TaskScheduler scheduler.Scheduler
		Backoff       backoff.Backoff
		NewTask       func(context.Context) (task tm.Task, args []interface{}, err error)
		TaskManager   func(context.Context) tm.TaskManger
	}

	//JobScheduler менеджер периодических задач
	JobScheduler interface {
		ID() string
		Subject() observer.Subject
		Enable(bool)
		Schedule()
		SetScheduler(sch scheduler.Scheduler)
		Close() error
	}
)

// NewJobScheduler создаем новую cron-задачу
func NewJobScheduler(appCtx context.Context, conf JobSchedulerConf) (JobScheduler, error) {
	const api = "NewJobScheduler"

	if conf.TaskManager == nil {
		t := tm.TaskManagerFromContext(appCtx)
		if t == nil {
			return nil, errors.Wrap(errors.New("'TaskManager' is not provided"), api)
		}
		conf.TaskManager = func(_ context.Context) tm.TaskManger {
			return t
		}
	}
	if conf.TaskScheduler == nil {
		return nil, errors.Wrap(errors.New("'TaskScheduler' is not provided"), api)
	}
	if conf.NewTask == nil {
		return nil, errors.Wrap(errors.New("'NewTask' is not provided"), api)
	}
	if conf.Backoff == nil {
		conf.Backoff = &backoff.StopBackoff
	}
	ret := &jobSchedulerImpl{
		subject:          observer.NewSubject(),
		appCtx:           appCtx,
		conf:             conf,
		closed:           make(chan struct{}),
		asyncEventsQueue: queue.NewFIFO(context.Background()),
	}
	subj := ret.subject
	que := ret.asyncEventsQueue
	go func() { //notify async events
		for {
			v, e := que.Get(context.Background())
			if e != nil {
				_ = que.Close()
				return
			}
			switch t := v.(type) {
			case closeEventQueue:
				_ = que.Close()
			case allow2continue:
				close(t)
			case observer.EventType:
				subj.Notify(t)
			}
		}
	}()
	return ret, nil
}

//---------------------------------===================== IMPL =====================---------------------------------

type (
	jobSchedulerImpl struct { //nolint:unused
		sync.Mutex
		appCtx           context.Context
		subject          observer.Subject
		enabled          int32
		closed           chan struct{}
		scheduleOnce     sync.Once
		closeOnce        sync.Once
		conf             JobSchedulerConf
		timer            *time.Timer
		lastSucceededAt  time.Time
		lastFinishedAt   time.Time
		runningJob       *taskWithArgs
		asyncEventsQueue queue.SimpleQueue
	}

	taskWithArgs struct {
		tm.CancellableTask
		args        []interface{}
		scheduledAt time.Time
		round       int32
		c           allow2continue
	}

	protectSubject struct {
		observer.Subject
	}

	allow2continue chan struct{}

	closeEventQueue struct{}
)

// DetachAllObservers override observer.Subject.DetachAllObservers
func (ps *protectSubject) DetachAllObservers() {}

func (man *jobSchedulerImpl) Close() error {
	var doClose bool
	man.closeOnce.Do(func() {
		man.scheduleOnce.Do(func() {})
		doClose = true
	})
	if doClose {
		man.asyncNotify(
			man.log("scheduler will close"),
			OnJobSchedulerClose{JobID: man.ID()},
		)
		close(man.closed)
		man.Lock()
		c, t := man.runningJob, man.timer
		man.timer = nil
		man.Unlock()
		if c != nil {
			c.Cancel()
		}
		if t != nil {
			_ = t.Stop()
		}
		man.asyncNotify(closeEventQueue{})
	}
	return nil
}

// Subject ...
func (man *jobSchedulerImpl) Subject() observer.Subject {
	return &protectSubject{Subject: man.subject}
}

// ID ...
func (man *jobSchedulerImpl) ID() string {
	return man.conf.JobID
}

func (man *jobSchedulerImpl) SetScheduler(sch scheduler.Scheduler) {
	if sch == nil {
		return
	}
	man.Lock()
	defer man.Unlock()
	man.conf.TaskScheduler = sch
	if man.timer == nil {
		return
	}
	if man.runningJob != nil {
		return
	}
	tme := man.calcPauseDuration(man.lastSucceededAt, false)
	man.timer.Stop()
	man.timer.Reset(tme)
}

// Schedule ...
func (man *jobSchedulerImpl) Schedule() {
	man.scheduleOnce.Do(func() {
		man.Lock()
		man.getBackoff().Reset()
		delta := man.calcPauseDuration(time.Time{}, false)
		man.Unlock()
		man.subject.Notify(
			man.log("scheduler is started; job will run in (%v)", delta),
			OnJobSchedulerStarted{JobID: man.ID()},
		)
		man.timer = time.AfterFunc(delta, man.runJob)
	})
}

// Enable enable or not enable
func (man *jobSchedulerImpl) Enable(enabled bool) {
	var stateChanged bool
	if enabled {
		stateChanged = atomic.CompareAndSwapInt32(&man.enabled, 0, 1)
	} else {
		stateChanged = atomic.CompareAndSwapInt32(&man.enabled, 1, 0)
	}
	if !stateChanged {
		return
	}

	man.Lock()
	defer man.Unlock()
	var logMsg string
	if !enabled {
		if man.timer != nil {
			man.timer.Stop()
		}
		logMsg = "caught 'disable' signal"
		if man.runningJob != nil {
			logMsg += "; current job will stop"
			man.runningJob.Cancel()
		}
	} else {
		logMsg = "caught 'enable' signal"
		if man.runningJob == nil {
			man.getBackoff().Reset()
			if man.timer != nil && !man.lastFinishedAt.IsZero() {
				sleepDuration := man.calcPauseDuration(man.lastSucceededAt, false)
				logMsg += fmt.Sprintf("; job will run in (%v)", sleepDuration)
				man.timer.Reset(sleepDuration)
			}
		}
	}
	man.asyncNotify(
		man.log(logMsg),
		OnJobSchedulerEnabled{
			JobID:   man.ID(),
			Enabled: enabled,
		},
	)
}

func (man *jobSchedulerImpl) isEnabled() bool {
	select {
	case <-man.appCtx.Done():
	case <-man.closed:
	default:
		return atomic.AddInt32(&man.enabled, 0) != 0
	}
	return false
}

func (man *jobSchedulerImpl) calcPauseDuration(startTime time.Time, fromBackoff bool) time.Duration {
	var ret time.Duration
	if fromBackoff {
		ret = man.getBackoff().NextBackOff()
	} else {
		ret = man.conf.TaskScheduler.NextActivity(startTime).Sub(time.Now()) //nolint:gosimple
	}
	if ret != backoff.Stop && ret < minDelayBeforeStart {
		ret = minDelayBeforeStart
	}
	return ret
}

func (man *jobSchedulerImpl) asyncNotify(evts ...interface{}) {
	for i := range evts {
		man.asyncEventsQueue.Put(evts[i])
	}
}

func (man *jobSchedulerImpl) runJob() {
	const (
		causeDisabled = "job could not run cause scheduler is disabled"
	)
	man.Lock()
	locked := true
	defer func() {
		if locked {
			man.Unlock()
		}
	}()
	if man.runningJob != nil {
		return
	}
	if !man.isEnabled() {
		man.asyncNotify(man.log(causeDisabled))
		return
	}

	man.Unlock()
	locked = false
	newJob, err := man.createNewJob()
	if err != nil {
		err = errors.Wrap(ErrNoTaskIsProvided, err.Error())
		man.asyncNotify(
			man.log("scheduler will stop by reason: %v", err),
			OnJobSchedulerStop{
				JobID:  man.ID(),
				Reason: err,
			},
		)
		return
	}
	taskManager := man.conf.TaskManager(man.appCtx)
	if taskManager == nil {
		newJob.Cancel()
		man.asyncNotify(
			man.log("scheduler will stop by reason: %v", ErrNoTaskManager),
			OnJobSchedulerStop{
				JobID:  man.ID(),
				Reason: ErrNoTaskManager,
			},
		)
		return
	}
	man.Lock()
	locked = true
	if !man.isEnabled() {
		newJob.Cancel()
		man.asyncNotify(man.log(causeDisabled))
		return
	}
	man.runningJob = newJob
	man.runningJob.scheduledAt = time.Now()
	man.asyncNotify(
		man.log("job is staring"),
		OnJobStarted{
			JobID: man.ID(),
			At:    man.runningJob.scheduledAt,
		},
	)
	newJob.c = make(allow2continue)
	_ = taskManager.Schedule(man.runningJob, func(_ tm.TaskInfo, taskResult []interface{}, failure error) {
		atomic.AddInt32(&man.runningJob.round, 1)
		man.finishJob(taskResult, failure)
	}, man.runningJob.args...)
}

func (man *jobSchedulerImpl) finishJob(jobResults []interface{}, jobStartFailure error) {
	now := time.Now()
	evt := OnJobFinished{
		JobID: man.ID(),
		At:    now,
	}
	evt.setResult(jobResults, jobStartFailure)
	failure := evt.FindError()

	man.Lock()
	wait4continue := man.runningJob.c
	man.lastFinishedAt = now
	man.runningJob.Cancel()
	man.runningJob = nil
	if failure == nil {
		man.lastSucceededAt = now
	}
	man.Unlock()

	if failure != nil {
		man.asyncNotify(man.log("job has failed; error: %v", failure))
	} else {
		man.asyncNotify(man.log("job has successfully finished"))
	}
	man.asyncNotify(evt, wait4continue)

	select {
	case <-wait4continue:
	case <-man.closed:
	case <-man.appCtx.Done():
	}
	if !man.isEnabled() {
		man.asyncNotify(man.log("scheduler won`t plan next run cause it is disabled"))
		return
	}

	man.Lock()
	defer man.Unlock()
	if failure == nil || errors.Is(failure, context.Canceled) {
		failure = nil
		man.getBackoff().Reset()
	}
	sleepDuration := man.calcPauseDuration(man.lastSucceededAt, failure != nil)
	if sleepDuration == backoff.Stop {
		man.asyncNotify(
			man.log("scheduler will stop by reason: %v", ErrBackoffStopped),
			OnJobSchedulerStop{
				JobID:  man.ID(),
				Reason: ErrBackoffStopped,
			})
		return
	}

	man.asyncNotify(
		man.log("scheduler will sleep (%v) until (%s)",
			sleepDuration,
			time.Now().Add(sleepDuration).Format(logTimeLayout)),
	)

	man.timer.Reset(sleepDuration)
}

func (man *jobSchedulerImpl) getBackoff() backoff.Backoff {
	return man.conf.Backoff
}

func (man *jobSchedulerImpl) log(format string, args ...interface{}) OnJobLog {
	return OnJobLog{
		TextMessageEvent: observer.NewTextEvent(format, args...),
		JobID:            man.ID(),
	}
}

func (man *jobSchedulerImpl) createNewJob() (*taskWithArgs, error) {
	ctx, cancel := context.WithCancel(man.appCtx)
	t, args, err := man.conf.NewTask(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	cancels := []func(){cancel}
	if c, ok := t.(tm.CancellableTask); ok {
		cancels = append(cancels, c.Cancel)
	}
	return &taskWithArgs{
		CancellableTask: tm.MakeCancellableTask(
			tm.OverrideTaskID(t, man.ID()),
			func() {
				for _, c := range cancels {
					c()
				}
			},
		),
		args: args,
	}, nil
}
