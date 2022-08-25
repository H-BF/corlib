package signals

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/H-BF/corlib/pkg/patterns/observer"
	"github.com/stretchr/testify/assert"
)

func Test_SubjectOfSignals(t *testing.T) {
	ch := make(chan SignalFromOS, 1)
	defer close(ch)
	obs := observer.NewObserver(func(ev observer.EventType) {
		select {
		case ch <- ev.(SignalFromOS):
		default:
		}
	}, false, SignalFromOS{})
	SubjOfSignalsFromOS().ObserversAttach(obs)
	p, _ := os.FindProcess(os.Getpid())
	e := p.Signal(syscall.SIGHUP)
	assert.NoError(t, e)
	if e != nil {
		return
	}

	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()

	select {
	case <-ch:
	case <-ctx.Done():
		e = ctx.Err()
	}
	assert.NoError(t, e)
}
