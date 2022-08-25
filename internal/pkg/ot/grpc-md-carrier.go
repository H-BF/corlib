package ot

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

//TextMapCarrierFromGrpcMD makes propagation.TextMapCarrier from metadata.MD
type TextMapCarrierFromGrpcMD struct {
	metadata.MD
}

var _ propagation.TextMapCarrier = (*TextMapCarrierFromGrpcMD)(nil)

//Get impl propagation.TextMapCarrier
func (impl TextMapCarrierFromGrpcMD) Get(key string) string {
	if values := impl.MD.Get(key); len(values) > 0 {
		return values[0]
	}
	return ""
}

//Set impl propagation.TextMapCarrier
func (impl TextMapCarrierFromGrpcMD) Set(key string, value string) {
	impl.MD.Set(key, value)
}

//Keys impl propagation.TextMapCarrier
func (impl TextMapCarrierFromGrpcMD) Keys() []string {
	out := make([]string, 0, impl.Len())
	for key := range impl.MD {
		out = append(out, key)
	}
	return out
}

//FillFromHTTPHeader fills GRPC metadata from HTTP Header
func (impl TextMapCarrierFromGrpcMD) FillFromHTTPHeader(hh http.Header) {
	prop := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	ctx := prop.Extract(context.Background(), propagation.HeaderCarrier(hh))
	prop.Inject(ctx, impl)
}
