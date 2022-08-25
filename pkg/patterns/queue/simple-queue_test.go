package queue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_QueueInConcurrency(t *testing.T) {
	ctx := context.Background()
	que := NewFIFO(ctx)
	var sum1, sum2 int64
	go func() {
		for i := 1; i <= 3000; i++ {
			sum1 += int64(i)
			que.Put(i)
			time.Sleep(time.Millisecond)
		}
		_ = que.Close()
	}()
	for {
		v, e := que.Get(ctx)
		if e != nil {
			assert.Equal(t, ErrQueueClosed, e)
			break
		}
		sum2 += int64(v.(int))
	}
	assert.Equal(t, sum1, sum2)
}

func Test_QueueFIFO(t *testing.T) {
	ctx := context.Background()
	que := NewFIFO(ctx)
	expected := []int{1, 2, 3, 4, 5}
	que.Put(1, 2, 3, 4, 5)
	_ = que.Close()
	var got []int
	for {
		v, e := que.Get(ctx)
		if e != nil {
			assert.Equal(t, ErrQueueClosed, e)
			break
		}
		got = append(got, v.(int))
	}
	assert.Equal(t, expected, got)
}

func Test_QueueLIFO(t *testing.T) {
	ctx := context.Background()
	que := NewLIFO(ctx)

	que.Put(1, 2, 3, 4, 5)
	_ = que.Close()

	expected := []int{5, 4, 3, 2, 1}
	var got []int

	for {
		v, e := que.Get(ctx)
		if e != nil {
			assert.Equal(t, ErrQueueClosed, e)
			break
		}
		got = append(got, v.(int))
	}

	assert.Equal(t, expected, got)
}
