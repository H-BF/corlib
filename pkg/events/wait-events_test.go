package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_WaitOne(t *testing.T) {
	cases := []struct {
		name string
		f    func(t2 *testing.T)
	}{
		{
			"fire-on-timeout",
			func(t2 *testing.T) {
				ctx := context.Background()
				e1 := NewEvent(1)
				e2 := NewEvent(2)
				_ = time.AfterFunc(time.Second, func() {
					e2.Fire()
				})
				what, err := WaitOne(ctx, e1, e2)
				assert.NoError(t2, err)
				assert.Equal(t2, 1, what)
			},
		},
		{
			"failed-on-deadline",
			func(t2 *testing.T) {
				ctx, c := context.WithTimeout(context.Background(), time.Second)
				defer c()
				e1 := NewEvent(1)
				e2 := NewEvent(2)
				_, err := WaitOne(ctx, e1, e2)
				assert.True(t2, errors.Is(err, context.DeadlineExceeded))
			},
		},
		{
			"cancelled",
			func(t2 *testing.T) {
				e1 := NewEvent(1)
				e2 := NewEvent(2)
				ctx, c := context.WithCancel(context.Background())
				_ = time.AfterFunc(time.Second, func() {
					c()
				})
				_, err := WaitOne(ctx, e1, e2)
				assert.True(t2, errors.Is(err, context.Canceled))
			},
		},
	}
	for _, v := range cases {
		t.Run(v.name, v.f)
	}
}

func Test_WaitAll(t *testing.T) {
	cases := []struct {
		name string
		f    func(t2 *testing.T)
	}{
		{
			"fire-on-timeout",
			func(t2 *testing.T) {
				ctx := context.Background()
				e1 := NewEvent(1)
				e2 := NewEvent(2)
				e3 := NewEvent(2)

				_ = time.AfterFunc(time.Second, func() {
					e1.Fire()
				})
				_ = time.AfterFunc(2*time.Second, func() {
					e2.Fire()
				})
				_ = time.AfterFunc(3*time.Second, func() {
					e3.Fire()
				})
				err := WaitAll(ctx, e1, e2, e3)
				assert.NoError(t2, err)
			},
		},
		{
			"failed-on-deadline",
			func(t2 *testing.T) {
				ctx, c := context.WithTimeout(context.Background(), time.Second)
				defer c()
				e1 := NewEvent(1)
				e2 := NewEvent(2)
				e3 := NewEvent(2)
				_ = time.AfterFunc(time.Second, func() {
					e1.Fire()
				})
				_ = time.AfterFunc(2*time.Second, func() {
					e2.Fire()
				})
				_ = time.AfterFunc(3*time.Second, func() {
					e3.Fire()
				})
				err := WaitAll(ctx, e1, e2, e3)
				assert.True(t2, errors.Is(err, context.DeadlineExceeded))
			},
		},
		{
			"cancelled",
			func(t2 *testing.T) {
				ctx, c := context.WithCancel(context.Background())
				_ = time.AfterFunc(time.Second, func() {
					c()
				})
				e1 := NewEvent(1)
				e2 := NewEvent(2)
				e3 := NewEvent(2)
				_ = time.AfterFunc(time.Second, func() {
					e1.Fire()
				})
				_ = time.AfterFunc(2*time.Second, func() {
					e2.Fire()
				})
				_ = time.AfterFunc(3*time.Second, func() {
					e3.Fire()
				})
				err := WaitAll(ctx, e1, e2, e3)
				assert.True(t2, errors.Is(err, context.Canceled))
			},
		},
	}
	for _, v := range cases {
		t.Run(v.name, v.f)
	}
}
