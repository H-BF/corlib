package signals

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/H-BF/corlib/pkg/patterns/observer"
)

// SignalFromOS signals incoming from OS
type SignalFromOS struct {
	observer.EventType
	syscall.Signal
}

// SubjOfSignalsFromOS signals incoming from OS subject
func SubjOfSignalsFromOS() observer.Subject {
	return subjOfSignalsFromOS
}

var subjOfSignalsFromOS = observer.NewSubject()

func init() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch)
	go func() {
		for {
			switch sig := (<-ch).(type) {
			case syscall.Signal:
				subjOfSignalsFromOS.Notify(SignalFromOS{Signal: sig})
			}
		}
	}()
}
