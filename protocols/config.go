// Package protocols defines common interfaces and types for all transport protocols
package protocols

import (
	"errors"
	"time"
)

// Config is the interface that all configuration types must implement
type Config interface {
	// Validate validates the configuration
	Validate() error
	// Merge merges this configuration with another configuration
	Merge(other Config) (Config, error)
	// Clone creates a deep copy of the configuration
	Clone() Config
}

// BaseConfig provides common configuration options for all protocols
type BaseConfig struct {
	// Timeout is the default timeout for operations
	Timeout time.Duration
	// MaxRetries is the maximum number of retries for failed operations
	MaxRetries int
	// Metadata is a map of additional configuration options
	Metadata map[string]any
}

// NewBaseConfig creates a new BaseConfig with default values
func NewBaseConfig() *BaseConfig {
	return &BaseConfig{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		Metadata:   make(map[string]any),
	}
}

// Validate implements Config interface
func (c *BaseConfig) Validate() error {
	if c.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	if c.MaxRetries < 0 {
		return errors.New("max retries must be non-negative")
	}
	return nil
}

// Merge implements Config interface
func (c *BaseConfig) Merge(other Config) (Config, error) {
	otherBase, ok := other.(*BaseConfig)
	if !ok {
		return nil, errors.New("cannot merge with non-base config")
	}

	merged := c.Clone().(*BaseConfig)
	if otherBase.Timeout > 0 {
		merged.Timeout = otherBase.Timeout
	}
	if otherBase.MaxRetries >= 0 {
		merged.MaxRetries = otherBase.MaxRetries
	}
	for k, v := range otherBase.Metadata {
		merged.Metadata[k] = v
	}

	return merged, nil
}

// Clone implements Config interface
func (c *BaseConfig) Clone() Config {
	clone := &BaseConfig{
		Timeout:    c.Timeout,
		MaxRetries: c.MaxRetries,
		Metadata:   make(map[string]any),
	}
	for k, v := range c.Metadata {
		clone.Metadata[k] = v
	}
	return clone
}

// GetTimeout returns the timeout duration
func (c *BaseConfig) GetTimeout() time.Duration {
	return c.Timeout
}

// GetMaxRetries returns the maximum number of retries
func (c *BaseConfig) GetMaxRetries() int {
	return c.MaxRetries
}

// SetTimeout sets the timeout duration
func (c *BaseConfig) SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

// SetMaxRetries sets the maximum number of retries
func (c *BaseConfig) SetMaxRetries(retries int) {
	c.MaxRetries = retries
}

// SetMetadata sets a metadata value
func (c *BaseConfig) SetMetadata(key string, value any) {
	c.Metadata[key] = value
}

// GetMetadata gets a metadata value
func (c *BaseConfig) GetMetadata(key string) (any, bool) {
	value, ok := c.Metadata[key]
	return value, ok
}
