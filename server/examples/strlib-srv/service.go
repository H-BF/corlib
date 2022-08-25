package strlib_srv

import (
	"context"
	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/server"
	"github.com/H-BF/corlib/server/tests"
	"github.com/H-BF/corlib/server/tests/strlib"
	"github.com/go-openapi/spec"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"strings"
)

//NewStrLibService ...
func NewStrLibService(ctx context.Context) *strLibSrv {
	return &strLibSrv{appCtx: ctx}
}

var (
	_ server.APIService = (*strLibSrv)(nil)
)

type strLibSrv struct {
	strlib.UnimplementedStrlibServer
	appCtx context.Context
}

func (src *strLibSrv) GetDocs() (*spec.Swagger, error) {
	return tests.GetStrlibDocs()
}

func (src *strLibSrv) Description() grpc.ServiceDesc {
	return strlib.Strlib_ServiceDesc
}

func (src *strLibSrv) RegisterGRPC(_ context.Context, grpcServer *grpc.Server) error {
	strlib.RegisterStrlibServer(grpcServer, src)
	return nil
}

func (src *strLibSrv) RegisterProxyGW(ctx context.Context, mux *runtime.ServeMux, c *grpc.ClientConn) error {
	return strlib.RegisterStrlibHandler(ctx, mux, c)
}

func (src *strLibSrv) OnStart() {

}

func (src *strLibSrv) OnStop() {

}

func (src *strLibSrv) Uppercase(ctx context.Context, req *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error) {
	const api = "Uppercase"
	logger.Debugf(ctx, "strlib.%s", api)
	var ret strlib.UppercaseResponse
	ret.Value = strings.ToUpper(req.GetValue())
	return &ret, nil
}
