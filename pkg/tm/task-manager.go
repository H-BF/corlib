package tm

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/H-BF/corlib/pkg/functional"
)

// TaskStatus TBD
type TaskStatus int8

const (
	//TaskStatusScheduled TBD
	TaskStatusScheduled TaskStatus = iota + 1
	//TaskStatusFinished TBD
	TaskStatusFinished
)

type (
	//TaskCompletion TDB
	TaskCompletion func(taskInfo TaskInfo, taskResult []interface{}, err error)

	//TaskControl TBD
	TaskControl struct {
		Wait4Completed func(ctx context.Context) (TaskState, error)
		QueryState     func() TaskState
	}

	//TaskManger TBD
	TaskManger interface {
		Schedule(t Task, completion TaskCompletion, args ...interface{}) TaskControl
		QueryTaskState(taskID TaskID) *TaskState
		Enum(func(TaskState) error) error
	}

	//TaskState TBD
	TaskState struct {
		TaskInfo
		Status      TaskStatus
		Err         error
		Args        []interface{}
		Result      []interface{}
		ScheduledAt *time.Time
		FinishedAt  *time.Time
	}

	internalTaskState struct {
		TaskState
		completionReceivers map[reflect.Value]struct{} //Just unique receivers
		barrier             chan struct{}
	}

	keyOfTask TaskID

	taskManagerTasks map[keyOfTask]*internalTaskState

	taskManager struct {
		sync.Mutex
		active taskManagerTasks
	}
)

// NewTaskManager creates TaskManger instance
func NewTaskManager() TaskManger {
	return &taskManager{
		active: make(taskManagerTasks),
	}
}

func (tm *taskManager) finishTask(key keyOfTask, taskResult []interface{}, err error) {
	var receivers map[reflect.Value]struct{}
	var taskInfo TaskInfo
	tm.Lock()
	state := tm.active[key]
	delete(tm.active, key)
	tm.Unlock()
	if state == nil {
		return
	}
	defer close(state.barrier)
	now := time.Now()
	taskInfo = state.TaskInfo
	state.FinishedAt = &now
	state.Err = err
	state.Result = taskResult
	if state.Status == TaskStatusScheduled {
		receivers = state.completionReceivers
	}
	state.completionReceivers, state.Status = nil, TaskStatusFinished
	for f := range receivers {
		callback := f.Interface().(TaskCompletion)
		callback(taskInfo, taskResult, err)
	}
}

func (tm *taskManager) Schedule(t Task, completion TaskCompletion, args ...interface{}) TaskControl {
	tm.Lock()
	defer tm.Unlock()
	key := keyOfTask(t.ID())
	if state := tm.active[key]; state != nil {
		if state.Status != TaskStatusFinished {
			if completion != nil {
				state.completionReceivers[reflect.ValueOf(completion)] = struct{}{}
			}
			return TaskControl{
				Wait4Completed: state.await4completed,
				QueryState: func() TaskState {
					tm.Lock()
					defer tm.Unlock()
					return state.TaskState
				},
			}
		}
		delete(tm.active, key)
	}
	timePoint := time.Now()
	newState := &internalTaskState{
		barrier: make(chan struct{}),
	}
	newState.TaskInfo = t
	newState.Status = TaskStatusScheduled
	newState.ScheduledAt = &timePoint
	newState.completionReceivers = map[reflect.Value]struct{}{}
	newState.Args = args
	if completion != nil {
		newState.completionReceivers[reflect.ValueOf(completion)] = struct{}{}
	}
	tm.active[key] = newState
	taskFinalize := func(result []interface{}, err error) {
		tm.finishTask(key, result, err)
	}
	_ = functional.MustAsyncJob(t.Invoke).
		WhenStateChanged(func(state functional.AsyncJobState) {
			switch state.Status {
			case functional.AsyncJobStartFailed:
				taskFinalize(nil, state.Failure)
			case functional.AsyncJobCompleted:
				taskResults, _ := state.Returned[0].([]interface{})
				taskError, _ := state.Returned[1].(error)
				taskFinalize(taskResults, taskError)
			}
		}).Run(args...)
	return TaskControl{
		Wait4Completed: newState.await4completed,
		QueryState: func() TaskState {
			tm.Lock()
			defer tm.Unlock()
			return newState.TaskState
		},
	}
}

func (tm *taskManager) QueryTaskState(taskID TaskID) *TaskState {
	tm.Lock()
	defer tm.Unlock()
	if state := tm.active[keyOfTask(taskID)]; state != nil {
		ret := state.TaskState
		return &ret
	}
	return nil
}

func (tm *taskManager) Enum(visitor func(TaskState) error) error {
	tm.Lock()
	defer tm.Unlock()
	for _, item := range tm.active {
		if e := visitor(item.TaskState); e != nil {
			return e
		}
	}
	return nil
}

func (ts *internalTaskState) await4completed(ctx context.Context) (TaskState, error) {
	var (
		ret TaskState
		err error
	)
	ret.TaskInfo = ts.TaskInfo
	ret.ScheduledAt = ts.ScheduledAt
	ret.Args = ts.Args
	select {
	case <-ts.barrier:
		ret = ts.TaskState
	case <-ctx.Done():
		err = ctx.Err()
	}
	return ret, err
}
