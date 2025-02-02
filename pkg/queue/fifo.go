package queue

import (
	"container/list"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// FIFO interface
type FIFO[T any] interface {
	Reader() <-chan T
	Put(v ...T) bool
	Len() int
	Close() error
}

// NewFIFO -
func NewFIFO[T any]() *typedFIFO[T] { //nolint:revive
	return &typedFIFO[T]{
		data:  list.New(),
		cv:    sync.NewCond(new(sync.Mutex)),
		close: make(chan struct{}),
		ch:    make(chan T),
	}
}

var _ FIFO[int] = (*typedFIFO[int])(nil)

type typedFIFO[T any] struct {
	data        *list.List
	close       chan struct{}
	stopped     chan struct{}
	ch          chan T
	cv          *sync.Cond
	sendPending uint32
	closeOnce   sync.Once
	runOnce     sync.Once
}

// Len impl FIFO[T] interface
func (que *typedFIFO[T]) Len() int {
	que.cv.L.Lock()
	defer que.cv.L.Unlock()
	if que.data != nil {
		return que.data.Len() + int(atomic.LoadUint32(&que.sendPending))
	}
	return 0
}

// Reader impl FIFO[T] interface
func (que *typedFIFO[T]) Reader() <-chan T {
	que.runOnce.Do(func() {
		que.stopped = make(chan struct{})
		go que.run()
	})
	return que.ch
}

// Put impl FIFO[T] interface
func (que *typedFIFO[T]) Put(v ...T) (ok bool) {
	que.cv.L.Lock()
	defer func() {
		if ok {
			que.cv.Broadcast()
		}
		que.cv.L.Unlock()
		if ok {
			runtime.Gosched()
		}
	}()
	if que.data != nil {
		for i := range v {
			que.data.PushBack(v[i])
		}
		ok = len(v) > 0
	}
	return ok
}

// Close impl FIFO[T] interface
func (que *typedFIFO[T]) Close() error {
	que.runOnce.Do(func() {})
	stopped := que.stopped
	cv := que.cv
	cl := que.close
	ch := que.ch
	que.closeOnce.Do(func() {
		const waitBeforeBroadcast = 100 * time.Millisecond
		close(cl)
		defer close(ch)
		cv.L.Lock()
		que.data = nil
		cv.L.Unlock()
		if stopped != nil {
		loop:
			for cv.Broadcast(); ; cv.Broadcast() {
				select {
				case <-stopped:
					break loop
				case <-time.After(waitBeforeBroadcast):
				}
			}
		}
	})
	return nil
}

func (que *typedFIFO[T]) run() {
	defer close(que.stopped)
	for closed := false; !closed; {
		if v, ok := que.fetch(); !ok {
			break
		} else {
			atomic.StoreUint32(&que.sendPending, 1)
			select {
			case <-que.close:
				closed = true
			case que.ch <- v.(T):
			}
			atomic.StoreUint32(&que.sendPending, 0)
		}
	}
}

func (que *typedFIFO[T]) fetch() (v any, ok bool) {
	que.cv.L.Lock()
	defer que.cv.L.Unlock()
	data := que.data
	for ; data != nil; data = que.data {
		if o := data.Front(); o != nil {
			v, ok = o.Value, true
			data.Remove(o)
			break
		}
		que.cv.Wait()
	}
	return v, ok
}
