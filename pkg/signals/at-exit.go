package signals

import (
	"context"
	"sync"
	"syscall"

	"github.com/H-BF/corlib/pkg/lazy"
	"github.com/H-BF/corlib/pkg/patterns/observer"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// WhenSignalExit adds `func() error` callback to the globalCloser
func WhenSignalExit(f ...func() error) {
	globalAtExitManager.Value().(*AtExitManager).WhenSignalExit(f...)
}

// AtExitManager ...
type AtExitManager struct {
	sync.Mutex
	closeOnce    sync.Once
	closed       chan struct{}
	rip          []func() error
	isClosing    bool
	obs          observer.Observer
	errFromClose error
}

// NewAtExitManager returns new WhenSignalExit, if []os.Signal is specified Closer will automatically
// call Close when one of signals is received from OS
func NewAtExitManager() *AtExitManager {
	return &AtExitManager{closed: make(chan struct{})}
}

// WhenSignalExit register RIP functions
func (c *AtExitManager) WhenSignalExit(f ...func() error) {
	c.Lock()
	defer c.Unlock()
	if c.isClosing {
		return
	}
	if len(c.rip) == 0 {
		c.obs = observer.NewObserver(func(event observer.EventType) {
			sig := event.(SignalFromOS)
			switch sig.Signal {
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGABRT:
				_ = c.Close()
			}
		}, true, SignalFromOS{})
		SubjOfSignalsFromOS().ObserversAttach(c.obs)
	}
	c.rip = append(c.rip, f...)
}

// Wait4Closed blocks until all closer functions are done
func (c *AtExitManager) Wait4Closed(ctx context.Context) error {
	select {
	case <-c.closed:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// Close calls all closer functions
func (c *AtExitManager) Close() error {
	const api = "AtExitManager.Close"
	c.closeOnce.Do(func() {
		SubjOfSignalsFromOS().ObserversDetach(c.obs)
		c.obs = nil
		defer close(c.closed)
		c.Lock()
		c.isClosing = true
		funcs := c.rip
		c.rip = nil
		c.Unlock()
		errs := make([]error, 0, len(funcs))
		for _, f := range funcs {
			if e := f(); e != nil {
				errs = append(errs, e)
			}
		}
		c.errFromClose = errors.Wrap(multierr.Combine(errs...), api)
	})
	return c.errFromClose
}

var (
	globalAtExitManager = lazy.MakeInitializer(func() interface{} {
		return NewAtExitManager()
	})
)
