// Package metrics provides metrics collection for transportx.
package metrics

import "time"

// Config holds configuration for metrics collection
type Config struct {
	// Connection metrics
	EnableConnections    bool
	EnableActiveStreams  bool
	EnableFailedConnects bool

	// Performance metrics
	EnableLatency    bool
	EnableThroughput bool
	EnableErrors     bool
	EnableRetries    bool

	// Resource metrics
	EnableMemoryUsage bool
	EnableGoroutines  bool

	// Thresholds
	MaxConnections    int64
	MaxActiveStreams  int64
	MaxFailedConnects int64
	MaxLatency        time.Duration
	MaxThroughput     int64
	MaxErrors         int64
	MaxRetries        int64
	MaxMemoryUsage    int64
	MaxGoroutines     int64
}

// NewConfig creates a new metrics configuration with default values
func NewConfig() *Config {
	return &Config{
		// Enable all metrics by default
		EnableConnections:    true,
		EnableActiveStreams:  true,
		EnableFailedConnects: true,
		EnableLatency:        true,
		EnableThroughput:     true,
		EnableErrors:         true,
		EnableRetries:        true,
		EnableMemoryUsage:    true,
		EnableGoroutines:     true,

		// Default thresholds
		MaxConnections:    1000,
		MaxActiveStreams:  100,
		MaxFailedConnects: 100,
		MaxLatency:        5 * time.Second,
		MaxThroughput:     10000,
		MaxErrors:         100,
		MaxRetries:        10,
		MaxMemoryUsage:    1024 * 1024 * 1024, // 1GB
		MaxGoroutines:     1000,
	}
}

// WithConnections enables/disables connection metrics
func (c *Config) WithConnections(enable bool) *Config {
	c.EnableConnections = enable
	return c
}

// WithActiveStreams enables/disables active streams metrics
func (c *Config) WithActiveStreams(enable bool) *Config {
	c.EnableActiveStreams = enable
	return c
}

// WithFailedConnects enables/disables failed connections metrics
func (c *Config) WithFailedConnects(enable bool) *Config {
	c.EnableFailedConnects = enable
	return c
}

// WithLatency enables/disables latency metrics
func (c *Config) WithLatency(enable bool) *Config {
	c.EnableLatency = enable
	return c
}

// WithThroughput enables/disables throughput metrics
func (c *Config) WithThroughput(enable bool) *Config {
	c.EnableThroughput = enable
	return c
}

// WithErrors enables/disables error metrics
func (c *Config) WithErrors(enable bool) *Config {
	c.EnableErrors = enable
	return c
}

// WithRetries enables/disables retry metrics
func (c *Config) WithRetries(enable bool) *Config {
	c.EnableRetries = enable
	return c
}

// WithMemoryUsage enables/disables memory usage metrics
func (c *Config) WithMemoryUsage(enable bool) *Config {
	c.EnableMemoryUsage = enable
	return c
}

// WithGoroutines enables/disables goroutine metrics
func (c *Config) WithGoroutines(enable bool) *Config {
	c.EnableGoroutines = enable
	return c
}

// WithMaxConnections sets the maximum connections threshold
func (c *Config) WithMaxConnections(limit int64) *Config {
	c.MaxConnections = limit
	return c
}

// WithMaxActiveStreams sets the maximum active streams threshold
func (c *Config) WithMaxActiveStreams(limit int64) *Config {
	c.MaxActiveStreams = limit
	return c
}

// WithMaxFailedConnects sets the maximum failed connections threshold
func (c *Config) WithMaxFailedConnects(limit int64) *Config {
	c.MaxFailedConnects = limit
	return c
}

// WithMaxLatency sets the maximum latency threshold
func (c *Config) WithMaxLatency(limit time.Duration) *Config {
	c.MaxLatency = limit
	return c
}

// WithMaxThroughput sets the maximum throughput threshold
func (c *Config) WithMaxThroughput(limit int64) *Config {
	c.MaxThroughput = limit
	return c
}

// WithMaxErrors sets the maximum errors threshold
func (c *Config) WithMaxErrors(limit int64) *Config {
	c.MaxErrors = limit
	return c
}

// WithMaxRetries sets the maximum retries threshold
func (c *Config) WithMaxRetries(limit int64) *Config {
	c.MaxRetries = limit
	return c
}

// WithMaxMemoryUsage sets the maximum memory usage threshold
func (c *Config) WithMaxMemoryUsage(limit int64) *Config {
	c.MaxMemoryUsage = limit
	return c
}

// WithMaxGoroutines sets the maximum goroutines threshold
func (c *Config) WithMaxGoroutines(limit int64) *Config {
	c.MaxGoroutines = limit
	return c
}
