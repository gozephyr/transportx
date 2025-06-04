// Package grpc implements a gRPC transport for the transportx library.
// It provides a client and server implementation for gRPC-based communication.
package grpc

// BatchMessage represents multiple messages
type BatchMessage struct {
	Messages []*Message
}

// BatchRequest represents a request for batch operations
type BatchRequest struct {
	BatchSize int32
}

// Empty represents an empty message
type Empty struct{}
