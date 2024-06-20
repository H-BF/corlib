package queue

import (
	"context"
	"errors"
)

// Type type of queue
type Type int

const (
	// TypeFIFO FIFO queue
	TypeFIFO Type = iota
	// TypeLIFO LIFO queue
	TypeLIFO
)

// Queue base interface
type Queue interface {
	Type() Type
	Close() error
}

// SimpleQueue simple FIFO | LIFO queue
type SimpleQueue interface {
	Queue
	Put(val ...interface{}) bool
	Get(ctx context.Context) (interface{}, error)
}

// ErrQueueClosed when queue is closed
var ErrQueueClosed = errors.New("queue is closed")
