package scheduler

import (
	"time"
)

//NewConstIntervalScheduler creates const interval scheduler
func NewConstIntervalScheduler(delta time.Duration) Scheduler {
	return &constIntervalScheduler{delta: delta}
}

// -----------------------------------------------------------------------------------------------------------------

type constIntervalScheduler struct {
	delta time.Duration
}

func (sh *constIntervalScheduler) Close() error {
	return nil
}

func (sh *constIntervalScheduler) NextActivity(startPoint time.Time) time.Time {
	now := time.Now()
	if startPoint.IsZero() {
		return now
	}
	delta := now.Sub(startPoint)
	if delta < 0 {
		return startPoint
	}
	if delta > sh.delta {
		return now
	}
	return startPoint.Add(sh.delta)
}
