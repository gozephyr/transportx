// Package grpc implements a gRPC transport for the transportx library.
// It provides a client and server implementation for gRPC-based communication.
package grpc

// Message represents a gRPC message with data payload
type Message struct {
	Data []byte
}
