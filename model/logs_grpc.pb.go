// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.10.0
// source: logs.proto

package model

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

// LogEntryServiceClient is the client API for LogEntryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LogEntryServiceClient interface {
	WriteLogEntry(ctx context.Context, in *WriteLogCommand, opts ...grpc.CallOption) (*LogEntryResult, error)
	GetLogEntry(ctx context.Context, in *LogEntryQuery, opts ...grpc.CallOption) (*LogEntryResult, error)
	GetLogEntries(ctx context.Context, in *LogEntriesQuery, opts ...grpc.CallOption) (*LogEntriesResult, error)
}

type logEntryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewLogEntryServiceClient(cc grpc.ClientConnInterface) LogEntryServiceClient {
	return &logEntryServiceClient{cc}
}

func (c *logEntryServiceClient) WriteLogEntry(ctx context.Context, in *WriteLogCommand, opts ...grpc.CallOption) (*LogEntryResult, error) {
	out := new(LogEntryResult)
	err := c.cc.Invoke(ctx, "/model.LogEntryService/WriteLogEntry", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *logEntryServiceClient) GetLogEntry(ctx context.Context, in *LogEntryQuery, opts ...grpc.CallOption) (*LogEntryResult, error) {
	out := new(LogEntryResult)
	err := c.cc.Invoke(ctx, "/model.LogEntryService/GetLogEntry", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *logEntryServiceClient) GetLogEntries(ctx context.Context, in *LogEntriesQuery, opts ...grpc.CallOption) (*LogEntriesResult, error) {
	out := new(LogEntriesResult)
	err := c.cc.Invoke(ctx, "/model.LogEntryService/GetLogEntries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LogEntryServiceServer is the server API for LogEntryService service.
// All implementations should embed UnimplementedLogEntryServiceServer
// for forward compatibility
type LogEntryServiceServer interface {
	WriteLogEntry(context.Context, *WriteLogCommand) (*LogEntryResult, error)
	GetLogEntry(context.Context, *LogEntryQuery) (*LogEntryResult, error)
	GetLogEntries(context.Context, *LogEntriesQuery) (*LogEntriesResult, error)
}

// UnimplementedLogEntryServiceServer should be embedded to have forward compatible implementations.
type UnimplementedLogEntryServiceServer struct {
}

func (UnimplementedLogEntryServiceServer) WriteLogEntry(context.Context, *WriteLogCommand) (*LogEntryResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WriteLogEntry not implemented")
}
func (UnimplementedLogEntryServiceServer) GetLogEntry(context.Context, *LogEntryQuery) (*LogEntryResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLogEntry not implemented")
}
func (UnimplementedLogEntryServiceServer) GetLogEntries(context.Context, *LogEntriesQuery) (*LogEntriesResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLogEntries not implemented")
}

// UnsafeLogEntryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LogEntryServiceServer will
// result in compilation errors.
type UnsafeLogEntryServiceServer interface {
	mustEmbedUnimplementedLogEntryServiceServer()
}

func RegisterLogEntryServiceServer(s grpc.ServiceRegistrar, srv LogEntryServiceServer) {
	s.RegisterService(&LogEntryService_ServiceDesc, srv)
}

func _LogEntryService_WriteLogEntry_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WriteLogCommand)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogEntryServiceServer).WriteLogEntry(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/model.LogEntryService/WriteLogEntry",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogEntryServiceServer).WriteLogEntry(ctx, req.(*WriteLogCommand))
	}
	return interceptor(ctx, in, info, handler)
}

func _LogEntryService_GetLogEntry_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogEntryQuery)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogEntryServiceServer).GetLogEntry(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/model.LogEntryService/GetLogEntry",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogEntryServiceServer).GetLogEntry(ctx, req.(*LogEntryQuery))
	}
	return interceptor(ctx, in, info, handler)
}

func _LogEntryService_GetLogEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogEntriesQuery)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogEntryServiceServer).GetLogEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/model.LogEntryService/GetLogEntries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogEntryServiceServer).GetLogEntries(ctx, req.(*LogEntriesQuery))
	}
	return interceptor(ctx, in, info, handler)
}

// LogEntryService_ServiceDesc is the grpc.ServiceDesc for LogEntryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LogEntryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "model.LogEntryService",
	HandlerType: (*LogEntryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "WriteLogEntry",
			Handler:    _LogEntryService_WriteLogEntry_Handler,
		},
		{
			MethodName: "GetLogEntry",
			Handler:    _LogEntryService_GetLogEntry_Handler,
		},
		{
			MethodName: "GetLogEntries",
			Handler:    _LogEntryService_GetLogEntries_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "logs.proto",
}
