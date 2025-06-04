// Package protocols defines common interfaces and types for all transport protocols
package protocols

import (
	"context"
	"errors"
	"io"
	"time"
)

// Protocol defines the common interface that all transport protocols must implement
type Protocol interface {
	// Core connection methods
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error

	// Basic data transfer methods (required for all protocols)
	Send(ctx context.Context, data []byte) error
	Receive(ctx context.Context) ([]byte, error)

	// Protocol-specific methods
	GetProtocolName() string
	GetProtocolVersion() string

	// Configuration methods
	SetTimeout(timeout time.Duration)
	GetTimeout() time.Duration
	SetMaxRetries(retries int)
	GetMaxRetries() int

	// State methods
	IsConnected() bool
	GetConnectionState() ConnectionState
}

// StreamProtocol extends Protocol with streaming capabilities
type StreamProtocol interface {
	Protocol

	// Streaming methods
	NewStream(ctx context.Context) (Stream, error)
	AcceptStream(ctx context.Context) (Stream, error)
}

// Stream represents a bidirectional stream
type Stream interface {
	io.ReadWriteCloser
	Context() context.Context
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

// BatchProtocol extends Protocol with batch operation capabilities
type BatchProtocol interface {
	Protocol

	// Batch operations
	BatchSend(ctx context.Context, dataItems [][]byte) error
	BatchReceive(ctx context.Context, batchSize int) ([][]byte, error)
}

// PubSubProtocol extends Protocol with publish/subscribe capabilities
type PubSubProtocol interface {
	Protocol

	// Pub/Sub methods
	Subscribe(ctx context.Context, topic string) error
	Unsubscribe(ctx context.Context, topic string) error
	Publish(ctx context.Context, topic string, data []byte) error
}

// ConnectionState represents the current state of a protocol connection
type ConnectionState struct {
	Connected    bool
	LastActivity time.Time
	Error        error
	Metadata     map[string]any
}

// ProtocolConfig contains common configuration options for all protocols
type ProtocolConfig struct {
	// Common settings
	Timeout     time.Duration
	MaxRetries  int
	BufferSize  int
	EnableTLS   bool
	TLSCertFile string
	TLSKeyFile  string
	Metadata    map[string]any

	// Connection settings
	KeepAlive            time.Duration
	MaxConnections       int
	IdleTimeout          time.Duration
	ReconnectDelay       time.Duration
	MaxReconnectAttempts int
}

// NewProtocolConfig returns a new ProtocolConfig with default values
func NewProtocolConfig() *ProtocolConfig {
	return &ProtocolConfig{
		Timeout:              30 * time.Second,
		MaxRetries:           3,
		BufferSize:           32 * 1024, // 32KB
		EnableTLS:            false,
		KeepAlive:            30 * time.Second,
		MaxConnections:       100,
		IdleTimeout:          60 * time.Second,
		ReconnectDelay:       1 * time.Second,
		MaxReconnectAttempts: 5,
		Metadata:             make(map[string]any),
	}
}

// Validate validates the configuration
func (c *ProtocolConfig) Validate() error {
	if c.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	if c.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}
	if c.BufferSize <= 0 {
		return errors.New("buffer size must be positive")
	}
	if c.MaxConnections <= 0 {
		return errors.New("max connections must be positive")
	}
	if c.IdleTimeout <= 0 {
		return errors.New("idle timeout must be positive")
	}
	if c.ReconnectDelay <= 0 {
		return errors.New("reconnect delay must be positive")
	}
	if c.MaxReconnectAttempts < 0 {
		return errors.New("max reconnect attempts cannot be negative")
	}
	return nil
}

// WithTimeout sets the timeout
func (c *ProtocolConfig) WithTimeout(d time.Duration) *ProtocolConfig {
	c.Timeout = d
	return c
}

// WithMaxRetries sets the maximum number of retries
func (c *ProtocolConfig) WithMaxRetries(n int) *ProtocolConfig {
	c.MaxRetries = n
	return c
}

// WithBufferSize sets the buffer size
func (c *ProtocolConfig) WithBufferSize(n int) *ProtocolConfig {
	c.BufferSize = n
	return c
}

// WithTLS enables TLS and sets the certificate files
func (c *ProtocolConfig) WithTLS(certFile, keyFile string) *ProtocolConfig {
	c.EnableTLS = true
	c.TLSCertFile = certFile
	c.TLSKeyFile = keyFile
	return c
}

// WithKeepAlive sets the keepalive duration
func (c *ProtocolConfig) WithKeepAlive(d time.Duration) *ProtocolConfig {
	c.KeepAlive = d
	return c
}

// WithMaxConnections sets the maximum number of connections
func (c *ProtocolConfig) WithMaxConnections(n int) *ProtocolConfig {
	c.MaxConnections = n
	return c
}

// WithIdleTimeout sets the idle timeout
func (c *ProtocolConfig) WithIdleTimeout(d time.Duration) *ProtocolConfig {
	c.IdleTimeout = d
	return c
}

// WithReconnectDelay sets the reconnect delay
func (c *ProtocolConfig) WithReconnectDelay(d time.Duration) *ProtocolConfig {
	c.ReconnectDelay = d
	return c
}

// WithMaxReconnectAttempts sets the maximum number of reconnect attempts
func (c *ProtocolConfig) WithMaxReconnectAttempts(n int) *ProtocolConfig {
	c.MaxReconnectAttempts = n
	return c
}

// WithMetadata sets a metadata value
func (c *ProtocolConfig) WithMetadata(key string, value any) *ProtocolConfig {
	if c.Metadata == nil {
		c.Metadata = make(map[string]any)
	}
	c.Metadata[key] = value
	return c
}
