// Package http provides the HTTP transport implementation for transportx.
package http

import "errors"

// HTTP-specific errors
var (
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrNotConnected       = errors.New("transport not connected")
	ErrConnectionFailed   = errors.New("connection failed")
	ErrSendFailed         = errors.New("send operation failed")
	ErrReceiveFailed      = errors.New("receive operation failed")
	ErrBatchSizeExceeded  = errors.New("batch size exceeds maximum allowed")
	ErrInvalidBatchSize   = errors.New("invalid batch size")
	ErrTimeout            = errors.New("operation timed out")
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")
	ErrServerError        = errors.New("server returned error")
	ErrInvalidResponse    = errors.New("invalid server response")
	ErrConnectionClosed   = errors.New("connection closed")
	ErrBufferFull         = errors.New("buffer is full")
	ErrInvalidAddress     = errors.New("invalid server address")
	ErrDNSResolution      = errors.New("DNS resolution failed")
	ErrPoolExhausted      = errors.New("connection pool exhausted")
	ErrInvalidState       = errors.New("invalid transport state")
)
