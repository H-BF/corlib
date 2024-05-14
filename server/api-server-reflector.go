package server

import (
	"strings"

	"google.golang.org/grpc"
	grpcReflection "google.golang.org/grpc/reflection"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type (
	protodescResolver struct {
		protodesc.Resolver
	}

	reflectionExtensionResolver struct {
		grpcReflection.ExtensionResolver
	}

	refGrpcSrv struct {
		grpcReflection.GRPCServer
	}
)

var ( //TODO: leave for experiments in future
	_ = (*protodescResolver)(nil)
	_ = (*reflectionExtensionResolver)(nil)
)

func addGRPCreflector(s grpcReflection.GRPCServer) {
	opts := grpcReflection.ServerOptions{
		Services: refGrpcSrv{
			GRPCServer: s,
		},
		/*//TODO: leave for experiments in future
		ExtensionResolver: reflectionExtensionResolver{
			ExtensionResolver: protoregistry.GlobalTypes,
		},
		DescriptorResolver: protodescResolver{
			Resolver: protoregistry.GlobalFiles,
		},
		*/
	}
	srv := grpcReflection.NewServer(opts)
	rpb.RegisterServerReflectionServer(s, srv)
}

// GetServiceInfo impl grpcReflection.GRPCServer
func (s refGrpcSrv) GetServiceInfo() map[string]grpc.ServiceInfo {
	var hide []string
	ret := s.GRPCServer.GetServiceInfo()
	for k := range ret {
		if i := strings.LastIndex(k, "/"); i >= 0 {
			k1 := k[i+1:]
			if _, found := ret[k1]; found {
				hide = append(hide, k)
			}
		}
	}
	for i := range hide { //hides nonstandart service names
		delete(ret, hide[i])
	}
	return ret
}

//                               -= reflectionExtensionResolver =-

// FindExtensionByName impl grpcReflection.ExtensionResolver
func (r reflectionExtensionResolver) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	if i := strings.LastIndex(string(field), "/"); i >= 0 {
		field = field[i+1:]
	}
	return r.ExtensionResolver.FindExtensionByName(field)
}

// FindExtensionByNumber impl grpcReflection.ExtensionResolver
func (r reflectionExtensionResolver) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	if i := strings.LastIndex(string(message), "/"); i >= 0 {
		message = message[i+1:]
	}
	return r.ExtensionResolver.FindExtensionByNumber(message, field)
}

// RangeExtensionsByMessage impl grpcReflection.ExtensionResolver
func (r reflectionExtensionResolver) RangeExtensionsByMessage(message protoreflect.FullName, f func(protoreflect.ExtensionType) bool) {
	if i := strings.LastIndex(string(message), "/"); i >= 0 {
		message = message[i+1:]
	}
	r.ExtensionResolver.RangeExtensionsByMessage(message, f)
}

//                                   -= protodescResolver =-

// FindFileByPath impl protodesc.Resolver
func (r protodescResolver) FindFileByPath(p string) (protoreflect.FileDescriptor, error) {
	return r.Resolver.FindFileByPath(p)
}

// FindDescriptorByName impl protodesc.Resolver
func (r protodescResolver) FindDescriptorByName(n protoreflect.FullName) (protoreflect.Descriptor, error) {
	if i := strings.LastIndex(string(n), "/"); i >= 0 {
		n = n[i+1:]
	}
	return r.Resolver.FindDescriptorByName(n)
}
