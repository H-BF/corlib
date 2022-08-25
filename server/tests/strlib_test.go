package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/H-BF/corlib/logger"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/server"
	"github.com/H-BF/corlib/server/tests/strlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test_ServiceAPI_WorksGood(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	cases := []struct {
		name string
		addr string
	}{
		{name: "net/tcp", addr: "tcp://127.0.0.1:7000"},
		{
			name: "net/unix",
			addr: "unix://" + path.Join("/tmp", fmt.Sprintf("test-%v-%v.socket", os.Getpid(), time.Now().Nanosecond())),
		},
	}
	for i := range cases {
		c := cases[i]
		ok := t.Run(c.name, func(t *testing.T) {
			testServiceAPIWorksGood(c.addr, t)
		})
		if !ok {
			return
		}
	}
}

func testServiceAPIWorksGood(addr string, t *testing.T) {
	logger.SetLevel(zap.InfoLevel)
	servAPI := &WithOnStartStopEvents{
		ServiceAPIWithMock: new(StrLibImpl),
	}
	var _ server.APIService = servAPI
	if s, _ := interface{}(servAPI).(server.APIGatewayProxy); !assert.NotNil(t, s) {
		return
	}
	var checkOnStart, checkOnStop bool
	servAPI.ProvideMock().On("Uppercase",
		mock.Anything, mock.Anything).
		Return(func(ctx context.Context, req *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error) {
			v := req.GetValue()
			return &strlib.UppercaseResponse{Value: strings.ToUpper(v)}, nil
		}).
		On("OnStart").
		Return(func() {
			checkOnStart = true
		}).On("OnStop").
		Return(func() {
			checkOnStop = true
		})

	endpoint, err := pkgNet.ParseEndpoint(addr)
	assert.NoError(t, err)
	docs, _ := GetStrlibDocs()
	var srv *server.APIServer
	srv, err = server.NewAPIServer(server.WithServices(servAPI), server.WithDocs(docs, ""))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	serviceGone := make(chan error, 1)
	ctx, cancelServer := context.WithTimeout(context.Background(), 30000*time.Second)
	defer func() {
		cancelServer()
		select {
		case err = <-serviceGone:
			assert.NoErrorf(t, err, "server stopped correctly")
			assert.Equal(t, true, checkOnStart)
			assert.Equal(t, true, checkOnStop)
		case <-time.After(30 * time.Second):
			assert.Fail(t, "server did not stopped after 30s")
		}
	}()

	go func() {
		defer close(serviceGone)
		err = srv.Run(ctx, endpoint, server.RunWithGracefulStop(10*time.Second))
		assert.NoError(t, err)
		if err != nil {
			serviceGone <- err
		}
	}()
	var gConn *grpc.ClientConn
	if endpoint.IsUnixDomain() {
		addr = endpoint.FQN()
	} else {
		addr = endpoint.String()
	}
	//gConn, err = grpc.DialContext(ctx, endpoint.String(), grpc.WithInsecure(), grpc.WithBlock())
	gConn, err = grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	assert.NoErrorf(t, err, "connect to GRPC server")
	if err != nil {
		return
	}
	defer gConn.Close() //nolint
	client := strlib.NewStrlibClient(gConn)
	req := strlib.UppercaseQuery{Value: "qwerty-a"}
	var resp *strlib.UppercaseResponse
	resp, err = client.Uppercase(ctx, &req)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	if !assert.Equal(t, "QWERTY-A", resp.GetValue()) {
		return
	}
	{ //HTTP request
		httpClient := pkgNet.UDS.EnrichClient(nil)
		req.Value = "qwas"
		var data []byte
		data, err = json.Marshal(&req)
		if !assert.NoError(t, err) {
			return
		}
		b := bytes.NewBuffer(data)
		var httpResp *http.Response

		var scheme string
		if endpoint.IsUnixDomain() {
			scheme = string(pkgNet.SchemeUnixHTTP)
		} else {
			scheme = string(pkgNet.SchemeHTTP)
		}
		httpResp, err = httpClient.Post(scheme+"://"+endpoint.String()+"/v1/uppercase", "application/json", b)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.NotNil(t, httpResp.Body) {
			return
		}
		resp = nil
		_ = json.NewDecoder(httpResp.Body).Decode(&resp)
		assert.Equal(t, "QWAS", resp.GetValue())
	}
}
