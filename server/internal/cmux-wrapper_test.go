package internal

import (
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
	"github.com/stretchr/testify/assert"
)

func Test_CMux_CloseCorrect(t *testing.T) {
	ep, err := pkgNet.ParseEndpoint("tcp://127.0.0.1:6000")
	if !assert.NoError(t, err) {
		return
	}
	var mx cmux.CMux
	mx, err = NewCMux(ep)
	if !assert.NoError(t, err) {
		return
	}
	hL := mx.Match(cmux.HTTP1())
	var (
		e1 error
		e2 error
		wg sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		e1 = mx.Serve()
	}()
	go func() {
		defer wg.Done()
		for {
			var c net.Conn
			if c, e2 = hL.Accept(); e2 != nil {
				break
			}
			_ = c.Close()
		}
	}()
	time.Sleep(300 * time.Millisecond)
	mx.Close()
	wg.Wait()
	assert.True(t, errors.Is(e2, cmux.ErrListenerClosed) || errors.Is(e2, cmux.ErrServerClosed))
	checkOk := false
	if opErr := (*net.OpError)(nil); errors.As(e1, &opErr) {
		s := opErr.Err.Error()
		checkOk = opErr.Op == "accept" &&
			strings.Contains(s, "closed") &&
			strings.Contains(s, "connection")
	}
	assert.True(t, checkOk)
}
