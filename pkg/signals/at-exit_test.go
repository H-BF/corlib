package signals

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_WhenSignalExit(t *testing.T) {
	done := make(chan struct{})
	WhenSignalExit(func() error {
		close(done)
		return nil
	})
	p, _ := os.FindProcess(os.Getpid())
	e := p.Signal(syscall.SIGINT)
	assert.NoError(t, e)
	if e != nil {
		return
	}
	var success bool
	select {
	case <-done:
		success = true
	case <-time.After(3 * time.Second):
	}
	assert.Equal(t, true, success)
}
