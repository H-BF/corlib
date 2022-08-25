package interceptors

import (
	"context"
	"time"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/conventions"
	"github.com/H-BF/corlib/pkg/jsonview"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	logCallMethods struct{}
	represent2log  struct {
		Service  string      `json:"service"`
		Method   string      `json:"method"`
		Duration interface{} `json:"duration"`
		Req      interface{} `json:"req,omitempty"`
		Resp     interface{} `json:"resp,omitempty"`
		Error    interface{} `json:"err,omitempty"`
	}
)

//LogServerAPI ...
var LogServerAPI logCallMethods

//Unary ...
func (logCallMethods) Unary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	timePoint := time.Now()
	resp, err := handler(ctx, req)
	log := logger.FromContext(ctx)

	doLog := (err != nil && log.Enabled(zap.ErrorLevel)) ||
		(err == nil && log.Enabled(zap.DebugLevel))

	if doLog {
		var mi conventions.GrpcMethodInfo
		if mi.Init(info.FullMethod) == nil {
			rep := represent2log{
				Service:  mi.ServiceFQN,
				Method:   mi.Method,
				Duration: jsonview.Marshaler(time.Since(timePoint)),
				Req:      jsonview.Marshaler(req),
			}
			const (
				msg     = "Unary/SERVER-API"
				details = "details"
			)
			if err == nil {
				rep.Resp = jsonview.Marshaler(resp)
				log.Debugw(msg, details, rep)
			} else {
				rep.Error = jsonview.Marshaler(err)
				log.Errorw(msg, details, rep)
			}
		}
	}
	return resp, err
}

//Stream ...
func (logCallMethods) Stream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	timePoint := time.Now()
	ctx := ss.Context()
	err := handler(srv, ss)

	log := logger.FromContext(ctx)
	doLog := (err != nil && log.Enabled(zap.ErrorLevel)) ||
		(err == nil && log.Enabled(zap.DebugLevel))

	if doLog {
		var mi conventions.GrpcMethodInfo
		if mi.Init(info.FullMethod) == nil {
			rep := represent2log{
				Service:  mi.ServiceFQN,
				Method:   mi.Method,
				Duration: jsonview.Marshaler(time.Since(timePoint)),
			}
			const (
				msg     = "Stream/SERVER-API"
				details = "details"
			)
			if err == nil {
				log.Debugw(msg, details, rep)
			} else {
				rep.Error = jsonview.Marshaler(err)
				log.Errorw(msg, details, rep)
			}
		}
	}
	return err
}
