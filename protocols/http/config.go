// Package http provides the HTTP transport implementation for transportx.
package http

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/gozephyr/transportx/protocols"
)

// Default buffer sizes and timeouts
const (
	defaultBufferSize = 32 * 1024   // 32KB
	maxBufferSize     = 1024 * 1024 // 1MB
)

// Config holds the configuration for HTTP transport
type Config struct {
	protocols.BaseConfig
	Protocol           string        // http or https
	ServerAddress      string        // hostname or IP without protocol and port
	Port               int           // server port
	MaxIdleConns       int           // maximum number of idle connections
	MaxConnsPerHost    int           // maximum number of connections per host
	IdleTimeout        time.Duration // idle connection timeout
	KeepAlive          time.Duration // keep-alive duration
	ResponseTimeout    time.Duration // response timeout
	MaxWaitDuration    time.Duration // maximum wait duration for connection
	HealthCheckDelay   time.Duration // health check delay
	EnableHTTP2        bool          // enable HTTP/2
	MaxHeaderBytes     int           // maximum header bytes
	DisableCompression bool          // disable compression
	DisableKeepAlives  bool          // disable keep-alives
}

// NewConfig creates a new HTTP configuration with default values
func NewConfig() *Config {
	baseConfig := protocols.NewBaseConfig()
	return &Config{
		BaseConfig:         *baseConfig,
		Protocol:           "http",
		ServerAddress:      "localhost",
		Port:               8080,
		MaxIdleConns:       100,
		MaxConnsPerHost:    100,
		IdleTimeout:        90 * time.Second,
		KeepAlive:          30 * time.Second,
		ResponseTimeout:    30 * time.Second,
		MaxWaitDuration:    5 * time.Second,
		HealthCheckDelay:   30 * time.Second,
		EnableHTTP2:        true,
		MaxHeaderBytes:     32 * 1024,
		DisableCompression: false,
		DisableKeepAlives:  false,
	}
}

// GetFullURL returns the complete server URL including protocol and port
func (c *Config) GetFullURL() string {
	if c.Port == 0 {
		return fmt.Sprintf("%s://%s", c.Protocol, c.ServerAddress)
	}
	return fmt.Sprintf("%s://%s:%d", c.Protocol, c.ServerAddress, c.Port)
}

// Validate validates the HTTP configuration
func (c *Config) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return err
	}
	if c.ServerAddress == "" {
		return errors.New("server address is required")
	}
	if _, err := url.Parse(c.ServerAddress); err != nil {
		return fmt.Errorf("invalid server address: %w", err)
	}
	if c.MaxIdleConns <= 0 {
		return errors.New("max idle connections must be positive")
	}
	if c.MaxConnsPerHost <= 0 {
		return errors.New("max connections per host must be positive")
	}
	if c.IdleTimeout <= 0 {
		return errors.New("idle timeout must be positive")
	}
	if c.KeepAlive <= 0 {
		return errors.New("keepalive must be positive")
	}
	if c.ResponseTimeout <= 0 {
		return errors.New("response timeout must be positive")
	}
	if c.MaxWaitDuration <= 0 {
		return errors.New("max wait duration must be positive")
	}
	if c.HealthCheckDelay <= 0 {
		return errors.New("health check delay must be positive")
	}
	if c.MaxHeaderBytes <= 0 || c.MaxHeaderBytes > maxBufferSize {
		return fmt.Errorf("max header bytes must be between 1 and %d", maxBufferSize)
	}
	return nil
}

// Merge implements Config interface
func (c *Config) Merge(other protocols.Config) (protocols.Config, error) {
	otherConfig, ok := other.(*Config)
	if !ok {
		return nil, errors.New("cannot merge with non-http config")
	}

	merged := c.Clone().(*Config)
	if otherConfig.ServerAddress != "" {
		merged.ServerAddress = otherConfig.ServerAddress
	}
	if otherConfig.MaxIdleConns > 0 {
		merged.MaxIdleConns = otherConfig.MaxIdleConns
	}
	if otherConfig.MaxConnsPerHost > 0 {
		merged.MaxConnsPerHost = otherConfig.MaxConnsPerHost
	}
	if otherConfig.IdleTimeout > 0 {
		merged.IdleTimeout = otherConfig.IdleTimeout
	}
	if otherConfig.KeepAlive > 0 {
		merged.KeepAlive = otherConfig.KeepAlive
	}
	if otherConfig.ResponseTimeout > 0 {
		merged.ResponseTimeout = otherConfig.ResponseTimeout
	}
	if otherConfig.MaxWaitDuration > 0 {
		merged.MaxWaitDuration = otherConfig.MaxWaitDuration
	}
	if otherConfig.HealthCheckDelay > 0 {
		merged.HealthCheckDelay = otherConfig.HealthCheckDelay
	}
	if otherConfig.MaxHeaderBytes > 0 {
		merged.MaxHeaderBytes = otherConfig.MaxHeaderBytes
	}
	merged.EnableHTTP2 = otherConfig.EnableHTTP2
	merged.DisableCompression = otherConfig.DisableCompression
	merged.DisableKeepAlives = otherConfig.DisableKeepAlives

	return merged, nil
}

// Clone implements Config interface
func (c *Config) Clone() protocols.Config {
	clone := &Config{
		BaseConfig:         *c.BaseConfig.Clone().(*protocols.BaseConfig),
		ServerAddress:      c.ServerAddress,
		MaxIdleConns:       c.MaxIdleConns,
		MaxConnsPerHost:    c.MaxConnsPerHost,
		IdleTimeout:        c.IdleTimeout,
		KeepAlive:          c.KeepAlive,
		ResponseTimeout:    c.ResponseTimeout,
		MaxWaitDuration:    c.MaxWaitDuration,
		HealthCheckDelay:   c.HealthCheckDelay,
		MaxHeaderBytes:     c.MaxHeaderBytes,
		EnableHTTP2:        c.EnableHTTP2,
		DisableCompression: c.DisableCompression,
		DisableKeepAlives:  c.DisableKeepAlives,
	}
	return clone
}

// WithTimeout sets the timeout
func (c *Config) WithTimeout(d time.Duration) *Config {
	c.Timeout = d
	return c
}

// WithMaxRetries sets the maximum number of retries
func (c *Config) WithMaxRetries(n int) *Config {
	c.MaxRetries = n
	return c
}

// WithServerAddress sets the server address
func (c *Config) WithServerAddress(addr string) *Config {
	c.ServerAddress = addr
	return c
}

// WithMaxIdleConns sets the maximum number of idle connections
func (c *Config) WithMaxIdleConns(n int) *Config {
	c.MaxIdleConns = n
	return c
}

// WithMaxConnsPerHost sets the maximum number of connections per host
func (c *Config) WithMaxConnsPerHost(n int) *Config {
	c.MaxConnsPerHost = n
	return c
}

// WithIdleTimeout sets the idle timeout
func (c *Config) WithIdleTimeout(d time.Duration) *Config {
	c.IdleTimeout = d
	return c
}

// WithKeepAlive sets the keepalive duration
func (c *Config) WithKeepAlive(d time.Duration) *Config {
	c.KeepAlive = d
	return c
}

// WithResponseTimeout sets the response timeout
func (c *Config) WithResponseTimeout(d time.Duration) *Config {
	c.ResponseTimeout = d
	return c
}

// WithMaxWaitDuration sets the maximum wait duration
func (c *Config) WithMaxWaitDuration(d time.Duration) *Config {
	c.MaxWaitDuration = d
	return c
}

// WithHealthCheckDelay sets the health check delay
func (c *Config) WithHealthCheckDelay(d time.Duration) *Config {
	c.HealthCheckDelay = d
	return c
}

// WithHTTP2 enables or disables HTTP/2
func (c *Config) WithHTTP2(enable bool) *Config {
	c.EnableHTTP2 = enable
	return c
}

// WithMaxHeaderBytes sets the maximum header bytes
func (c *Config) WithMaxHeaderBytes(n int) *Config {
	c.MaxHeaderBytes = n
	return c
}

// WithCompression enables or disables compression
func (c *Config) WithCompression(enable bool) *Config {
	c.DisableCompression = !enable
	return c
}

// WithKeepAlives enables or disables keep-alives
func (c *Config) WithKeepAlives(enable bool) *Config {
	c.DisableKeepAlives = !enable
	return c
}

// WithMetadata sets metadata
func (c *Config) WithMetadata(key string, value any) *Config {
	c.BaseConfig.SetMetadata(key, value)
	return c
}

// SetMetadata sets a metadata value
func (c *Config) SetMetadata(key string, value any) {
	c.Metadata[key] = value
}
