package queue

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_FIFO_Len(t *testing.T) {
	que := NewFIFO[int]()
	que.Put(10)
	n := que.Len()
	require.Equal(t, 1, n)
	_ = que.Reader()
	time.Sleep(time.Second)
	n = que.Len()
	require.Equal(t, 1, n)
	que.Close()
}

func Test_FIFO(t *testing.T) {
	f := NewFIFO[any]()

	var exp []any
	var got []any
	exp = append(exp, 1, 2, "3")
	ok := f.Put(exp...)
	require.True(t, ok)
	r := f.Reader()
	rd := func() (ret any) {
		select {
		case ret = <-r:
		case <-time.After(time.Millisecond):
		}
		return ret
	}
	got = append(got, rd(), rd(), rd())
	require.Equal(t, exp, got)
	var wg sync.WaitGroup
	var nW, nR uint64
	wg.Add(1)
	go func() {
		defer wg.Done()
		for ok := true; ok; {
			if ok = f.Put(1); ok {
				atomic.AddUint64(&nW, 1)
			}
		}
	}()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ok := true; ok; {
				if _, ok = <-r; ok {
					atomic.AddUint64(&nR, 1)
				}
			}
		}()
	}
	time.Sleep(500 * time.Millisecond)
	_ = f.Close()
	wg.Wait()
	ok = f.Put(1)
	require.False(t, ok)
	_, ok = <-r
	require.False(t, ok)
	require.True(t, nW >= nR)
}
