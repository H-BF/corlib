package scheduler

import (
	"time"

	"github.com/H-BF/corlib/pkg/patterns/observer"
)

type (
	//Scheduler base interface
	Scheduler interface {
		NextActivity(fromTime time.Time) time.Time
		Close() error
	}

	//AsyncScheduler async capability
	AsyncScheduler interface {
		Scheduler
		Enable(bool)
		ConnectionPoint() observer.Subject
	}

	//OnActivate scheduler event when it activates by time-table
	OnActivate struct {
		observer.EventType
		When     time.Time
		UserData interface{}
	}
)
