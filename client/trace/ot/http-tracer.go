package ot

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	appIdentity "github.com/H-BF/corlib/app/identity"
	"github.com/H-BF/corlib/logger"
	netPkg "github.com/H-BF/corlib/pkg/net"

	"go.opentelemetry.io/otel/attribute"
	otCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// WrapClient добавляем клиету OTel трэйсинг
func WrapClient(c *http.Client, traceProv ClientTracerProvider) *http.Client {
	ret := new(http.Client)
	if c == nil {
		*ret = *http.DefaultClient
	} else {
		*ret = *c
	}
	transport := ret.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	ret.Transport = &transportWrapper{
		propagator:     propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}),
		RoundTripper:   transport,
		tracerProvider: traceProv,
	}
	return ret
}

type transportWrapper struct {
	propagator     propagation.TextMapPropagator
	tracerProvider ClientTracerProvider
	http.RoundTripper
}

// RoundTrip impl http.RoundTripper
func (tr *transportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	span, req2 := tr.spanStart(req)
	ctx, header := req2.Context(), req2.Header
	tr.propagator.Inject(ctx, propagation.HeaderCarrier(header))
	resp, err := tr.RoundTripper.RoundTrip(req2)
	tr.endSpan(span, resp, err)
	return resp, err
}

func (tr *transportWrapper) endSpan(span trace.Span, resp *http.Response, err error) {
	if span == nil {
		return
	}
	defer span.End()
	if err != nil {
		span.RecordError(err)
	} else {
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)
		span.SetAttributes(attrs...)
		if resp.StatusCode/100 == 2 {
			span.SetStatus(otCodes.Ok, "")
		} else {
			span.SetStatus(semconv.SpanStatusFromHTTPStatusCode(resp.StatusCode))
		}
	}
}

func (tr *transportWrapper) attrsFromRequest(req *http.Request) []attribute.KeyValue {
	anURL := req.URL
	scheme := anURL.Scheme
	var attrs []attribute.KeyValue
	if netPkg.SchemeUnixHTTP.Is(anURL.Scheme) {
		attrs = append(attrs,
			semconv.NetTransportUnix,
			semconv.HTTPTargetKey.String(path.Join(anURL.Host, anURL.RequestURI())),
		)
	} else {
		if len(scheme) == 0 {
			scheme = string(netPkg.SchemeHTTP)
		}
		attrs = append(attrs,
			semconv.NetTransportTCP,
			semconv.HTTPHostKey.String(anURL.Host),
			semconv.HTTPTargetKey.String(anURL.RequestURI()),
		)
	}
	attrs = append(attrs, semconv.HTTPSchemeKey.String(scheme))
	if len(req.Method) > 0 {
		attrs = append(attrs, semconv.HTTPMethodKey.String(req.Method))
	} else {
		attrs = append(attrs, semconv.HTTPMethodKey.String(http.MethodGet))
	}
	if username, _, ok := req.BasicAuth(); ok {
		attrs = append(attrs, semconv.ContainerIDKey.String(username))
	}
	if ua := req.UserAgent(); len(ua) > 0 {
		attrs = append(attrs, semconv.HTTPUserAgentKey.String(ua))
	}
	if req.ContentLength > 0 {
		attrs = append(attrs, semconv.HTTPRequestContentLengthKey.Int64(req.ContentLength))
	}

	{
		flavor := ""
		if req.ProtoMajor == 1 {
			flavor = fmt.Sprintf("1.%d", req.ProtoMinor)
		} else if req.ProtoMajor == 2 {
			flavor = "2"
		}
		if flavor != "" {
			attrs = append(attrs, semconv.HTTPFlavorKey.String(flavor))
		}
	}
	if logger.IsLevelEnabled(req.Context(), zap.DebugLevel) {
		userinfo := anURL.User
		if userinfo != nil {
			anURL.User = url.UserPassword("***", "***")
		}
		attrs = append(attrs, semconv.HTTPURLKey.String(anURL.String()))
		anURL.User = userinfo
	}
	return attrs
}

func (tr *transportWrapper) spanStart(req *http.Request) (trace.Span, *http.Request) {
	var span trace.Span
	if tp := tr.tracerProvider; tp != nil {
		name := req.URL.Path
		if netPkg.SchemeUnixHTTP.Is(req.URL.Scheme) {
			name = path.Join(req.URL.Host, req.URL.Path)
		}
		attrs := tr.attrsFromRequest(req)
		tracer := tp.Tracer(path.Join(appIdentity.Name, "http-client"))
		ctx := req.Context()
		ctx, span = tracer.Start(
			ctx,
			name,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attrs...),
		)
		req = req.WithContext(ctx)
	}
	return span, req
}
