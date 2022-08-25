package ot

import (
	"context"

	appIdentity "github.com/H-BF/corlib/app/identity"
	"go.opentelemetry.io/otel/attribute"
	sdkRes "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

//MakeAppResource make app trace e resource
func MakeAppResource(ctx context.Context) (*sdkRes.Resource, error) {
	//TODO: Maybe detect k8s options

	opts := []sdkRes.Option{
		sdkRes.WithOS(),
		sdkRes.WithOSType(),
		sdkRes.WithProcess(),
	}

	attrs := []attribute.KeyValue{
		semconv.ServiceInstanceIDKey.String(appIdentity.InstanceID.String()),
	}
	if len(appIdentity.Name) > 0 {
		attrs = append(attrs, semconv.ServiceNameKey.String(appIdentity.Name))
	}
	if len(appIdentity.Namespace) > 0 {
		attrs = append(attrs, semconv.ServiceNamespaceKey.String(appIdentity.Namespace))
	}
	if len(appIdentity.Version) > 0 {
		attrs = append(attrs, semconv.ServiceVersionKey.String(appIdentity.Version))
	}

	if len(appIdentity.BuildTS) > 0 {
		attrs = append(attrs, AppBuildTsKey.String(appIdentity.BuildTS))
	}
	if len(appIdentity.BuildBranch) > 0 {
		attrs = append(attrs, AppBuildBranchKey.String(appIdentity.BuildBranch))
	}
	if len(appIdentity.BuildHash) > 0 {
		attrs = append(attrs, AppBuildHashKey.String(appIdentity.BuildHash))
	}
	if len(appIdentity.BuildTag) > 0 {
		attrs = append(attrs, AppBuildTagKey.String(appIdentity.BuildTag))
	}

	if len(attrs) > 0 {
		opts = append(opts, sdkRes.WithAttributes(attrs...))
	}

	return sdkRes.New(ctx, opts...)
}
