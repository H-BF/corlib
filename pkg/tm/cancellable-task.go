package tm

import (
	"runtime"
	"sync"
)

// CancellableTask отменяемая задача
type CancellableTask interface {
	Task
	Cancel()
}

// MakeCancellableTask создать отменяемую задачу
func MakeCancellableTask(task Task, canceller func()) CancellableTask {
	ret := &cancellableTaskImpl{
		Task:   task,
		cancel: canceller,
	}
	runtime.SetFinalizer(ret, func(o *cancellableTaskImpl) {
		o.cancel()
	})
	return ret
}

type cancellableTaskImpl struct {
	Task
	once   sync.Once
	cancel func()
}

func (t *cancellableTaskImpl) Cancel() {
	if t.cancel != nil {
		t.once.Do(t.cancel)
	}
}
