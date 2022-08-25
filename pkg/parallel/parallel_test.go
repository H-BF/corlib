package parallel

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//gid Get goroutine ID
func gid() uint64 { //Get goroutine ID
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func Test_ExecAbstract(t *testing.T) {
	const taskCount = 16
	for concurrency := 0; concurrency < 20; concurrency++ {
		taskCounter := map[uint64]int{}
		var mutex sync.Mutex
		_ = ExecAbstract(taskCount, int32(concurrency), func(n int) error {
			time.Sleep(time.Millisecond)
			mutex.Lock()
			taskCounter[gid()]++
			mutex.Unlock()
			return nil
		})
		calcTaskCount := 0
		t.Logf("on concurrency: %v", concurrency)
		for k, v := range taskCounter {
			calcTaskCount += v
			t.Logf("\t\tGoroutine:%v -- TaskCount:%+v\n", k, v)
		}
		require.Equal(t, taskCount, calcTaskCount,
			fmt.Sprintf("parallel.ExecAbstract check on concurrency(%v)", concurrency))
	}
}

func Test_MapExecAbstract(t *testing.T) {
	const taskCount = 16
	makeTaskMapper := func(reducer func(int) error) TaskMapper {
		c := make(chan int)
		go func() {
			defer close(c)
			for i := 0; i < taskCount; i++ {
				c <- i
			}
		}()
		return func() func() error {
			if i, ok := <-c; ok {
				return func() error {
					return reducer(i)
				}
			}
			return nil
		}
	}

	var mutex sync.Mutex
	for concurrency := 0; concurrency <= taskCount; concurrency++ {
		taskCounter := make(map[uint64]int)
		p := 0
		mapper := makeTaskMapper(func(i int) error {
			time.Sleep(time.Millisecond)
			mutex.Lock()
			taskCounter[gid()]++
			p += i
			mutex.Unlock()
			return nil
		})
		err := MapExecAbstract(mapper, int32(concurrency))
		calcTaskCount := 0
		t.Logf("on concurrency: %v", concurrency)
		require.NoError(t, err)
		for k, v := range taskCounter {
			calcTaskCount += v
			t.Logf("\t\tGoroutine:%v -- TaskCount:%+v\n", k, v)
		}
		require.Equal(t, taskCount, calcTaskCount,
			fmt.Sprintf("parallel.MapExecAbstract check on concurrency(%v)", concurrency))

		require.Equal(t, (taskCount-1)*taskCount/2, p,
			fmt.Sprintf("parallel.MapExecAbstract check on concurrency(%v)", concurrency))
	}
}
