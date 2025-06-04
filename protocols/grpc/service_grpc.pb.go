// Package grpc implements a gRPC transport for the transportx library.
// It provides a client and server implementation for gRPC-based communication.
package grpc

import (
	"context"

	"google.golang.org/grpc"
)

// TransportServiceClient is the client API for TransportService service.
type TransportServiceClient interface {
	// Send sends a single message
	Send(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Message, error)
	// Receive receives a single message
	Receive(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Message, error)
	// BatchSend sends multiple messages
	BatchSend(ctx context.Context, in *BatchMessage, opts ...grpc.CallOption) (*Empty, error)
	// BatchReceive receives multiple messages
	BatchReceive(ctx context.Context, in *BatchRequest, opts ...grpc.CallOption) (*BatchMessage, error)
}

type transportServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTransportServiceClient(cc grpc.ClientConnInterface) TransportServiceClient {
	return &transportServiceClient{cc}
}

func (c *transportServiceClient) Send(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Message, error) {
	out := new(Message)
	err := c.cc.Invoke(ctx, "/grpc.TransportService/Send", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *transportServiceClient) Receive(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Message, error) {
	out := new(Message)
	err := c.cc.Invoke(ctx, "/grpc.TransportService/Receive", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *transportServiceClient) BatchSend(ctx context.Context, in *BatchMessage, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/grpc.TransportService/BatchSend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *transportServiceClient) BatchReceive(ctx context.Context, in *BatchRequest, opts ...grpc.CallOption) (*BatchMessage, error) {
	out := new(BatchMessage)
	err := c.cc.Invoke(ctx, "/grpc.TransportService/BatchReceive", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TransportServiceServer is the server API for TransportService service.
type TransportServiceServer interface {
	// Send sends a single message
	Send(context.Context, *Message) (*Message, error)
	// Receive receives a single message
	Receive(context.Context, *Empty) (*Message, error)
	// BatchSend sends multiple messages
	BatchSend(context.Context, *BatchMessage) (*Empty, error)
	// BatchReceive receives multiple messages
	BatchReceive(context.Context, *BatchRequest) (*BatchMessage, error)
}

// UnimplementedTransportServiceServer can be embedded to have forward compatible implementations.
type UnimplementedTransportServiceServer struct{}

func (s *UnimplementedTransportServiceServer) Send(context.Context, *Message) (*Message, error) {
	return nil, ErrStreamUnimplemented
}

func (s *UnimplementedTransportServiceServer) Receive(context.Context, *Empty) (*Message, error) {
	return nil, ErrStreamUnimplemented
}

func (s *UnimplementedTransportServiceServer) BatchSend(context.Context, *BatchMessage) (*Empty, error) {
	return nil, ErrStreamUnimplemented
}

func (s *UnimplementedTransportServiceServer) BatchReceive(context.Context, *BatchRequest) (*BatchMessage, error) {
	return nil, ErrStreamUnimplemented
}

// RegisterTransportServiceServer registers the server with the gRPC server
func RegisterTransportServiceServer(s *grpc.Server, srv TransportServiceServer) {
	s.RegisterService(&_TransportService_serviceDesc, srv)
}

//nolint:staticcheck
var _TransportService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.TransportService",
	HandlerType: (*TransportServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Send",
			Handler:    _TransportService_Send_Handler,
		},
		{
			MethodName: "Receive",
			Handler:    _TransportService_Receive_Handler,
		},
		{
			MethodName: "BatchSend",
			Handler:    _TransportService_BatchSend_Handler,
		},
		{
			MethodName: "BatchReceive",
			Handler:    _TransportService_BatchReceive_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service.proto",
}

//nolint:staticcheck
func _TransportService_Send_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransportServiceServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.TransportService/Send",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(TransportServiceServer).Send(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

//nolint:staticcheck
func _TransportService_Receive_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransportServiceServer).Receive(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.TransportService/Receive",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(TransportServiceServer).Receive(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

//nolint:staticcheck
func _TransportService_BatchSend_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(BatchMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransportServiceServer).BatchSend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.TransportService/BatchSend",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(TransportServiceServer).BatchSend(ctx, req.(*BatchMessage))
	}
	return interceptor(ctx, in, info, handler)
}

//nolint:staticcheck
func _TransportService_BatchReceive_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(BatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransportServiceServer).BatchReceive(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.TransportService/BatchReceive",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(TransportServiceServer).BatchReceive(ctx, req.(*BatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}
