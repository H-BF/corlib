package graceful_shutdown

import (
	"context"
)

type (
	//Status graceful shutdown-er run status
	Status = uint

	//TaskCompletion is a type alias
	TaskCompletion = <-chan struct{}

	//Task interface ueses in graceful shutdown pattern
	Task interface {
		Exec(context.Context) TaskCompletion
	}
)

const (
	//Completed when all tasks are done
	Completed Status = iota

	//Timeout when time deatline has exceeded
	Timeout
)
