// Package grpc implements a gRPC transport for the transportx library.
// It provides a client and server implementation for gRPC-based communication.
package grpc

import (
	"context"
	"fmt"
	"net"

	generated "github.com/gozephyr/transportx/proto/generated/proto"
	"google.golang.org/grpc"
)

// Server represents a gRPC server
type Server struct {
	generated.UnimplementedTransportServiceServer
	server *grpc.Server
	addr   string
}

// NewServer creates a new gRPC server
func NewServer(addr string) *Server {
	return &Server{
		server: grpc.NewServer(),
		addr:   addr,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	generated.RegisterTransportServiceServer(s.server, s)
	return s.server.Serve(lis)
}

// Stop stops the gRPC server
func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// Send implements the Send RPC method
func (s *Server) Send(ctx context.Context, req *generated.Message) (*generated.Message, error) {
	return &generated.Message{Data: req.GetData()}, nil
}

// Receive implements the Receive RPC method
func (s *Server) Receive(ctx context.Context, req *generated.Empty) (*generated.Message, error) {
	return &generated.Message{Data: []byte("test response")}, nil
}

// BatchSend implements the BatchSend RPC method
func (s *Server) BatchSend(ctx context.Context, req *generated.BatchMessage) (*generated.Empty, error) {
	return &generated.Empty{}, nil
}

// BatchReceive implements the BatchReceive RPC method
func (s *Server) BatchReceive(ctx context.Context, req *generated.BatchRequest) (*generated.BatchMessage, error) {
	messages := make([]*generated.Message, req.GetBatchSize())
	for i := int32(0); i < req.GetBatchSize(); i++ {
		messages[i] = &generated.Message{Data: []byte(fmt.Sprintf("test response %d", i))}
	}
	return &generated.BatchMessage{Messages: messages}, nil
}

// StreamMessages implements the bidirectional streaming RPC
func (s *Server) StreamMessages(stream generated.TransportService_StreamMessagesServer) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		// Echo the message back
		if err := stream.Send(&generated.Message{Data: msg.GetData()}); err != nil {
			return err
		}
	}
}
