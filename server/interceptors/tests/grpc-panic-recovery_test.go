package tests

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/conventions"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/pkg/parallel"
	"github.com/H-BF/corlib/server"
	"github.com/H-BF/corlib/server/interceptors"
	"github.com/H-BF/corlib/server/tests/strlib"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_Recovery(t *testing.T) {
	ep, err := pkgNet.ParseEndpoint("127.0.0.1:7200")
	if !assert.NoError(t, err) {
		return
	}
	logger.SetLevel(zap.InfoLevel)
	buf := bytes.NewBuffer(nil)
	loggerSink := io.MultiWriter(os.Stdout, buf)
	l := logger.NewWithSink(zap.InfoLevel, loggerSink)
	logger.SetLogger(l)
	bone := &fishBone{endPt: ep}
	errPanic := errors.New("P-A-N-I-C")
	var caughtErr error
	bone.v.Store(func(_ context.Context, _ *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error) { //nolint:unparam
		panic(errPanic)
	})
	bone.serverOptions = []server.APIServerOption{server.WithRecovery(
		interceptors.NewRecovery(interceptors.RecoveryWithHandler(
			func(ctx context.Context, _ conventions.GrpcMethodInfo, v interface{}) error {
				caughtErr, _ = v.(error)
				return nil
			})),
	)}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var srv *server.APIServer
	srv, err = bone.newServer()
	if !assert.NoError(t, err) {
		return
	}
	runners := []func() error{
		func() error {
			e := srv.Run(ctx, bone.endPt)
			assert.NoError(t, e)
			return e
		},
		func() error {
			defer cancel()
			client, e := bone.client4GRPC(ctx)
			if !assert.NoError(t, e) {
				return e
			}
			defer client.Close() //nolint
			r := &strlib.UppercaseQuery{Value: "abc"}
			_, _ = client.Uppercase(ctx, r)
			return nil
		},
	}
	nRunners := len(runners)
	err = parallel.ExecAbstract(nRunners, int32(nRunners)-1, func(i int) error {
		return runners[i]()
	})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, true, errors.Is(caughtErr, errPanic))
}
