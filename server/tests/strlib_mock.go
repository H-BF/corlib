package tests

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/H-BF/corlib/server"
	"github.com/H-BF/corlib/server/tests/strlib"
	"github.com/go-openapi/spec"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type (
	//ServiceAPIWithMock to test to
	ServiceAPIWithMock interface {
		server.APIService
		server.APIGatewayProxy
		ProvideMock() *mock.Mock
	}

	//StrLibImpl to test to
	StrLibImpl struct {
		strlib.UnimplementedStrlibServer
		mock.Mock
	}

	//WithOnStartStopEvents to test to
	WithOnStartStopEvents struct {
		ServiceAPIWithMock
	}
)

var (
	//go:embed strlib/strlib.swagger.json
	strLibRawSwagger []byte
	//GetStrlibDocs ...
	GetStrlibDocs = func() (*server.SwaggerSpec, error) {
		const api = "GetSwagger"

		ret := new(spec.Swagger)
		if err := json.Unmarshal(strLibRawSwagger, ret); err != nil {
			return nil, errors.Wrap(err, api)
		}
		return ret, nil
	}
	_ = GetStrlibDocs
)

//ProvideMock ...
func (sLib *StrLibImpl) ProvideMock() *mock.Mock {
	return &sLib.Mock
}

//Description ...
func (sLib *StrLibImpl) Description() grpc.ServiceDesc {
	return strlib.Strlib_ServiceDesc
}

//RegisterGRPC ...
func (sLib *StrLibImpl) RegisterGRPC(_ context.Context, srv *grpc.Server) error {
	strlib.RegisterStrlibServer(srv, sLib)
	return nil
}

//RegisterProxyGW ...
func (sLib *StrLibImpl) RegisterProxyGW(ctx context.Context, mux *runtime.ServeMux, c *grpc.ClientConn) error {
	return strlib.RegisterStrlibHandler(ctx, mux, c)
}

//Uppercase mocked
func (sLib *StrLibImpl) Uppercase(_a0 context.Context, _a1 *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error) {
	ret := sLib.Called(_a0, _a1)

	type f = func(context.Context, *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error)
	if rf, ok := ret.Get(0).(f); ok {
		return rf(_a0, _a1)
	}

	var r0 *strlib.UppercaseResponse
	if rf, ok := ret.Get(0).(func(context.Context, *strlib.UppercaseQuery) *strlib.UppercaseResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*strlib.UppercaseResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *strlib.UppercaseQuery) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

//OnStart ...
func (s *WithOnStartStopEvents) OnStart() {
	ret := s.ProvideMock().Called()
	if f, ok := ret.Get(0).(func()); ok {
		f()
	}
}

//OnStop ...
func (s *WithOnStartStopEvents) OnStop() {
	ret := s.ProvideMock().Called()
	if f, ok := ret.Get(0).(func()); ok {
		f()
	}
}

var (
	_ server.APIService   = (*StrLibImpl)(nil)
	_ strlib.StrlibServer = (*StrLibImpl)(nil)
	_ ServiceAPIWithMock  = (*StrLibImpl)(nil)
	_ ServiceAPIWithMock  = (*WithOnStartStopEvents)(nil)
)
