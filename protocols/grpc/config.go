// Package grpc implements a gRPC transport for the transportx library.
// It provides a client and server implementation for gRPC-based communication.
package grpc

import (
	"errors"
	"strconv"
	"time"

	"github.com/gozephyr/transportx/protocols"
	"google.golang.org/grpc"
)

// Config holds the configuration for the gRPC transport
type Config struct {
	// ServerAddress is the address of the gRPC server
	ServerAddress string

	// Port is the port number for the gRPC server
	Port int

	// Timeout is the timeout for gRPC operations
	Timeout time.Duration

	// MaxRetries is the maximum number of retries for failed operations
	MaxRetries int

	// MaxConcurrentStreams is the maximum number of concurrent streams per connection
	MaxConcurrentStreams uint32

	// InitialWindowSize is the initial window size for flow control
	InitialWindowSize int32

	// MaxHeaderListSize is the maximum size of header list
	MaxHeaderListSize uint32

	// KeepAliveTime is the time between keepalive pings
	KeepAliveTime time.Duration

	// KeepAliveTimeout is the timeout for keepalive pings
	KeepAliveTimeout time.Duration

	// TLSConfig holds TLS configuration if using secure connections
	TLSConfig *TLSConfig

	// DialOptions are additional gRPC dial options
	DialOptions []grpc.DialOption
}

// TLSConfig holds TLS configuration for secure gRPC connections
type TLSConfig struct {
	// CertFile is the path to the client certificate file
	CertFile string

	// KeyFile is the path to the client key file
	KeyFile string

	// CAFile is the path to the CA certificate file
	CAFile string

	// ServerName is the server name for TLS verification
	ServerName string

	// InsecureSkipVerify skips TLS verification if true
	InsecureSkipVerify bool
}

// NewConfig creates a new gRPC configuration with default values
func NewConfig() *Config {
	return &Config{
		ServerAddress:        "localhost",
		Port:                 50051,
		Timeout:              30 * time.Second,
		MaxRetries:           3,
		MaxConcurrentStreams: 100,
		InitialWindowSize:    1024 * 1024, // 1MB
		MaxHeaderListSize:    8192,        // 8KB
		KeepAliveTime:        30 * time.Second,
		KeepAliveTimeout:     10 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ServerAddress == "" {
		return ErrInvalidServerAddress
	}
	if c.Port <= 0 || c.Port > 65535 {
		return ErrInvalidPort
	}
	if c.Timeout <= 0 {
		return ErrInvalidTimeout
	}
	if c.MaxRetries < 0 {
		return ErrInvalidMaxRetries
	}
	if c.MaxConcurrentStreams == 0 {
		return ErrInvalidMaxConcurrentStreams
	}
	if c.InitialWindowSize <= 0 {
		return ErrInvalidInitialWindowSize
	}
	if c.MaxHeaderListSize == 0 {
		return ErrInvalidMaxHeaderListSize
	}
	if c.KeepAliveTime <= 0 {
		return ErrInvalidKeepAliveTime
	}
	if c.KeepAliveTimeout <= 0 {
		return ErrInvalidKeepAliveTimeout
	}
	return nil
}

// GetFullAddress returns the full server address including port
func (c *Config) GetFullAddress() string {
	return c.ServerAddress + ":" + strconv.Itoa(c.Port)
}

// WithTLS sets TLS configuration
func (c *Config) WithTLS(tlsConfig *TLSConfig) *Config {
	c.TLSConfig = tlsConfig
	return c
}

// WithDialOptions adds gRPC dial options
func (c *Config) WithDialOptions(opts ...grpc.DialOption) *Config {
	c.DialOptions = append(c.DialOptions, opts...)
	return c
}

// Clone creates a deep copy of the Config struct to satisfy protocols.Config interface
func (c *Config) Clone() protocols.Config {
	copyCfg := *c
	if c.TLSConfig != nil {
		copyTLS := *c.TLSConfig
		copyCfg.TLSConfig = &copyTLS
	}
	if c.DialOptions != nil {
		copyCfg.DialOptions = make([]grpc.DialOption, len(c.DialOptions))
		copy(copyCfg.DialOptions, c.DialOptions)
	}
	return &copyCfg
}

// Merge merges this configuration with another configuration
func (c *Config) Merge(other protocols.Config) (protocols.Config, error) {
	otherCfg, ok := other.(*Config)
	if !ok {
		return nil, errors.New("cannot merge with non-gRPC config")
	}
	merged := c.Clone().(*Config)
	if otherCfg.ServerAddress != "" {
		merged.ServerAddress = otherCfg.ServerAddress
	}
	if otherCfg.Port != 0 {
		merged.Port = otherCfg.Port
	}
	if otherCfg.Timeout != 0 {
		merged.Timeout = otherCfg.Timeout
	}
	if otherCfg.MaxRetries != 0 {
		merged.MaxRetries = otherCfg.MaxRetries
	}
	if otherCfg.MaxConcurrentStreams != 0 {
		merged.MaxConcurrentStreams = otherCfg.MaxConcurrentStreams
	}
	if otherCfg.InitialWindowSize != 0 {
		merged.InitialWindowSize = otherCfg.InitialWindowSize
	}
	if otherCfg.MaxHeaderListSize != 0 {
		merged.MaxHeaderListSize = otherCfg.MaxHeaderListSize
	}
	if otherCfg.KeepAliveTime != 0 {
		merged.KeepAliveTime = otherCfg.KeepAliveTime
	}
	if otherCfg.KeepAliveTimeout != 0 {
		merged.KeepAliveTimeout = otherCfg.KeepAliveTimeout
	}
	if otherCfg.TLSConfig != nil {
		copyTLS := *otherCfg.TLSConfig
		merged.TLSConfig = &copyTLS
	}
	if otherCfg.DialOptions != nil {
		merged.DialOptions = make([]grpc.DialOption, len(otherCfg.DialOptions))
		copy(merged.DialOptions, otherCfg.DialOptions)
	}
	return merged, nil
}
