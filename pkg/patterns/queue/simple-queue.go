package queue

import (
	"container/list"
	"context"
	"math"
	"runtime"
	"sync"
	"time"
)

//NewLIFO new LIFO queue
func NewLIFO(ctx context.Context) SimpleQueue {
	return newSimpleQueue2(ctx, true)
}

//NewFIFO new LIFO queue
func NewFIFO(ctx context.Context) SimpleQueue {
	return newSimpleQueue2(ctx, false)
}

//=====================================================================================================================

func newSimpleQueue2(ctx context.Context, isLifo bool) *simpleQueue {
	ret := &simpleQueue{
		lifo:   isLifo,
		mx:     new(sync.Mutex),
		data:   list.New(),
		notify: make(chan struct{}),
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ret.ctx, ret.close = context.WithTimeout(ctx, time.Duration(math.MaxInt64))
	runtime.SetFinalizer(ret, func(o *simpleQueue) {
		o.close()
		close(o.notify)
	})
	return ret
}

type simpleQueue struct {
	lifo   bool
	mx     *sync.Mutex
	data   *list.List
	notify chan struct{}
	ctx    context.Context
	close  func()
}

//Type impl SimpleQueue
func (imp *simpleQueue) Type() Type {
	if imp.lifo {
		return TypeLIFO
	}
	return TypeFIFO
}

//Put impl SimpleQueue
func (imp *simpleQueue) Put(vals ...interface{}) bool {
	if len(vals) > 0 {
		imp.mx.Lock()
		var closed bool
		defer func() {
			imp.mx.Unlock()
			if !closed {
				select {
				case imp.notify <- struct{}{}:
				default:
				}
			}
		}()
		select {
		case <-imp.ctx.Done():
			closed = true
			return false
		default:
			for i := range vals {
				imp.data.PushBack(vals[i])
			}
		}
	}
	return true
}

//Get impl SimpleQueue
func (imp *simpleQueue) Get(ctx context.Context) (interface{}, error) {
	for {
		imp.mx.Lock()
		val, ok := imp.fetch()
		var err error
		var isClosed bool
		if !ok {
			imp.mx.Unlock()
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-imp.ctx.Done():
				isClosed = true
			case <-imp.notify:
			}
			imp.mx.Lock()
			val, ok = imp.fetch()
		}
		imp.mx.Unlock()
		if ok {
			return val, nil
		}
		if err != nil {
			return nil, err
		}
		if isClosed {
			return nil, ErrQueueClosed
		}
	}
}

//Close impl Closer
func (imp *simpleQueue) Close() error {
	imp.close()
	return nil
}

func (imp *simpleQueue) fetch() (d interface{}, ok bool) {
	var el *list.Element
	if imp.lifo {
		el = imp.data.Back()
	} else {
		el = imp.data.Front()
	}
	if ok = el != nil; ok {
		d = el.Value
		imp.data.Remove(el)
	}
	return
}
