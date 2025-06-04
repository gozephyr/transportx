// Package errors provides error types for transportx.
package errors

import "errors"

var (
	// ErrUnsupportedTransport is returned when the transport type is not supported
	ErrUnsupportedTransport = errors.New("unsupported transport type")
	// ErrInvalidConfig is returned when the configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")
	// ErrNotConnected is returned when the transport is not connected
	ErrNotConnected = errors.New("transport not connected")
	// ErrConnectionFailed is returned when the connection fails
	ErrConnectionFailed = errors.New("connection failed")
	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timed out")
	// ErrBufferFull is returned when the buffer is full
	ErrBufferFull = errors.New("buffer is full")
	// ErrBufferEmpty is returned when the buffer is empty
	ErrBufferEmpty = errors.New("buffer is empty")
	// ErrInvalidData is returned when the data is invalid
	ErrInvalidData = errors.New("invalid data")
	// ErrMaxRetriesExceeded is returned when maximum retries are exceeded
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")
)
