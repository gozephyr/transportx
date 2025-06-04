// Package grpc implements a gRPC transport for the transportx library.
// It provides a client and server implementation for gRPC-based communication.
package grpc

import "errors"

// Error constants
var (
	ErrInvalidServerAddress        = errors.New("invalid server address")
	ErrInvalidPort                 = errors.New("invalid port number")
	ErrInvalidTimeout              = errors.New("invalid timeout")
	ErrInvalidMaxRetries           = errors.New("invalid max retries")
	ErrInvalidMaxConcurrentStreams = errors.New("invalid max concurrent streams")
	ErrInvalidInitialWindowSize    = errors.New("invalid initial window size")
	ErrInvalidMaxHeaderListSize    = errors.New("invalid max header list size")
	ErrInvalidKeepAliveTime        = errors.New("invalid keepalive time")
	ErrInvalidKeepAliveTimeout     = errors.New("invalid keepalive timeout")
	ErrConnectionFailed            = errors.New("connection failed")
	ErrSendFailed                  = errors.New("send failed")
	ErrReceiveFailed               = errors.New("receive failed")
	ErrStreamClosed                = errors.New("stream closed")
	ErrInvalidMessage              = errors.New("invalid message")
	ErrInvalidStreamState          = errors.New("invalid stream state")
	ErrStreamTimeout               = errors.New("stream timeout")
	ErrStreamCancelled             = errors.New("stream cancelled")
	ErrStreamReset                 = errors.New("stream reset")
	ErrStreamUnavailable           = errors.New("stream unavailable")
	ErrStreamExhausted             = errors.New("stream exhausted")
	ErrStreamLimitExceeded         = errors.New("stream limit exceeded")
	ErrStreamDeadlineExceeded      = errors.New("stream deadline exceeded")
	ErrStreamResourceExhausted     = errors.New("stream resource exhausted")
	ErrStreamUnimplemented         = errors.New("stream unimplemented")
	ErrStreamInternal              = errors.New("stream internal error")
	ErrStreamDataLoss              = errors.New("stream data loss")
	ErrStreamFailedPrecondition    = errors.New("stream failed precondition")
	ErrStreamAborted               = errors.New("stream aborted")
	ErrStreamOutOfRange            = errors.New("stream out of range")
)
