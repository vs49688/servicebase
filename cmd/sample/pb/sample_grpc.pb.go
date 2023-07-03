// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: sample.proto

package pb

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Teapot_AmIATeapot_FullMethodName = "/sample.Teapot/AmIATeapot"
)

// TeapotClient is the client API for Teapot service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TeapotClient interface {
	AmIATeapot(ctx context.Context, in *AmIATeapotRequest, opts ...grpc.CallOption) (*AmIATeapotResponse, error)
}

type teapotClient struct {
	cc grpc.ClientConnInterface
}

func NewTeapotClient(cc grpc.ClientConnInterface) TeapotClient {
	return &teapotClient{cc}
}

func (c *teapotClient) AmIATeapot(ctx context.Context, in *AmIATeapotRequest, opts ...grpc.CallOption) (*AmIATeapotResponse, error) {
	out := new(AmIATeapotResponse)
	err := c.cc.Invoke(ctx, Teapot_AmIATeapot_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TeapotServer is the server API for Teapot service.
// All implementations must embed UnimplementedTeapotServer
// for forward compatibility
type TeapotServer interface {
	AmIATeapot(context.Context, *AmIATeapotRequest) (*AmIATeapotResponse, error)
	mustEmbedUnimplementedTeapotServer()
}

// UnimplementedTeapotServer must be embedded to have forward compatible implementations.
type UnimplementedTeapotServer struct {
}

func (UnimplementedTeapotServer) AmIATeapot(context.Context, *AmIATeapotRequest) (*AmIATeapotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AmIATeapot not implemented")
}
func (UnimplementedTeapotServer) mustEmbedUnimplementedTeapotServer() {}

// UnsafeTeapotServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TeapotServer will
// result in compilation errors.
type UnsafeTeapotServer interface {
	mustEmbedUnimplementedTeapotServer()
}

func RegisterTeapotServer(s grpc.ServiceRegistrar, srv TeapotServer) {
	s.RegisterService(&Teapot_ServiceDesc, srv)
}

func _Teapot_AmIATeapot_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AmIATeapotRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TeapotServer).AmIATeapot(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Teapot_AmIATeapot_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TeapotServer).AmIATeapot(ctx, req.(*AmIATeapotRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Teapot_ServiceDesc is the grpc.ServiceDesc for Teapot service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Teapot_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "sample.Teapot",
	HandlerType: (*TeapotServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AmIATeapot",
			Handler:    _Teapot_AmIATeapot_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sample.proto",
}
