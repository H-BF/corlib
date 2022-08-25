package tm

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/H-BF/corlib/pkg/functional"
	"github.com/stretchr/testify/assert"
)

func Test_TaskManager(t *testing.T) {
	tm := NewTaskManager()
	taskDescr := "TASK-1"

	if true {
		var invoked bool
		task1, _ := MakeSimpleTask(taskDescr, func(a int, b string) (int, string) {
			time.Sleep(3 * time.Second)
			invoked = true
			return a + 10000, fmt.Sprintf("%s-R-%v", b, a)
		})
		tc := tm.Schedule(task1, nil, 1, "A")
		ctx := context.Background()
		state, _ := tc.Wait4Completed(ctx)
		assert.NoError(t, state.Err, "no error")
		assert.True(t, invoked, "is invoked")
		return
	}

	if true {
		const test1 = "simple job call"
		task1, _ := MakeSimpleTask(taskDescr, func(a int, b string) (int, string) {
			return a + 10000, fmt.Sprintf("%s-R-%v", b, a)
		})
		var success bool
		recv := func(a int, b string) {
			assert.Equal(t, 10001, a, test1)
			assert.Equal(t, "A-R-1", b, test1)
			success = true
		}
		ch := make(chan interface{})
		_ = tm.Schedule(task1, func(info TaskInfo, res []interface{}, err error) {
			defer close(ch)
			if err == nil {
				_, _ = functional.MustCallableOf(recv).Invoke(res...)
			}
		}, 1, "A")
		select {
		case <-ch:
		case <-time.After(20 * time.Second):
			t.Fatal(test1)
		}
		assert.Equal(t, true, success, test1)
	}

	if true {
		task2, _ := MakeSimpleTask(taskDescr, func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ch := make(chan interface{}, 1)
		_ = tm.Schedule(task2, func(_ TaskInfo, result []interface{}, err error) {
			if err != nil {
				ch <- err
			} else {
				ch <- struct{}{}
			}
		}, ctx)
		success := false
		select {
		case t := <-ch:
			switch t.(type) {
			case error:
			default:
				success = true
			}
		case <-time.After(20 * time.Second):
		}
		assert.Equal(t, true, success, "simple job called and cancelled")
	}

	if true {
		calledTask := make(map[int]struct{})
		mx := sync.Mutex{}
		wg := sync.WaitGroup{}
		recv := functional.MustCallableOf(func(b int) {
			mx.Lock()
			defer mx.Unlock()
			calledTask[b] = struct{}{}
		})
		task2, _ := MakeSimpleTask(taskDescr, func(n int) int {
			return n
		})
		task1, _ := MakeSimpleTask(taskDescr, func(n int) int {
			wg.Add(1)
			tm.Schedule(task2, func(_ TaskInfo, res []interface{}, err error) {
				defer wg.Done()
				if err == nil {
					_ = recv.InvokeNoResult(res...)
				}
			}, 2)
			return n
		})
		task1 = OverrideTaskID(task1, string(task2.ID()))
		wg.Add(1)
		tm.Schedule(task1, func(_ TaskInfo, res []interface{}, err error) {
			defer wg.Done()
			if err == nil {
				_ = recv.InvokeNoResult(res...)
			}
		}, 1)
		wg.Wait()
		assert.Equal(t, 1, len(calledTask), "schedule-same-task-ID-twice")
	}

}
