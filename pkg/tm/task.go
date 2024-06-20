package tm

import (
	"github.com/H-BF/corlib/pkg/functional"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type (
	//TaskID TBD
	TaskID = string

	//TaskInfo RBD
	TaskInfo interface {
		String() string
		ID() TaskID
	}

	//Task TBD
	Task interface {
		TaskInfo
		functional.Callable
	}

	taskInfo struct {
		id          TaskID
		description string
	}

	simpleTask struct {
		*taskInfo
		functional.Callable
	}

	overrideTaskID struct {
		Task
		id TaskID
	}
)

// MakeSimpleTask TBD
func MakeSimpleTask(description string, functionalObject interface{}) (Task, error) {
	const api = "MakeSimpleTask"
	callable, err := functional.MayCallableOf(functionalObject)
	if err != nil {
		return nil, errors.Wrap(err, api)
	}
	return &simpleTask{
		taskInfo: MakeTaskInfo(description),
		Callable: callable,
	}, nil
}

// MakeTaskInfo TBD
func MakeTaskInfo(descr string) *taskInfo { //nolint
	return &taskInfo{
		id:          uuid.NewV4().String(),
		description: descr,
	}
}

// OverrideTaskID TBD
func OverrideTaskID(task Task, id string) Task {
	return &overrideTaskID{
		Task: task,
		id:   id,
	}
}

func (t *taskInfo) ID() TaskID {
	return t.id
}

func (t *taskInfo) String() string {
	return t.description
}

func (t *overrideTaskID) ID() TaskID {
	return t.id
}
