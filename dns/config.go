package dns

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// ResolverConfig holds configuration for the DNS resolver
type ResolverConfig struct {
	// Base settings
	Timeout       time.Duration
	MaxRetries    int
	Metadata      map[string]any
	EnableIPv6    bool
	EnableIPv4    bool
	RotateServers bool // Whether to rotate through nameservers

	// Cache settings
	TTL             time.Duration
	MaxCacheSize    int
	CleanupInterval time.Duration
	NegativeTTL     time.Duration // TTL for caching negative responses

	// Nameserver settings
	Nameservers []string
	Port        uint16 // DNS server port

	// Statistics settings
	MaxStatsSize int // Maximum number of lookup times to track
}

// DefaultConfig returns default resolver configuration
func DefaultConfig() *ResolverConfig {
	return &ResolverConfig{
		Timeout:         30 * time.Second,
		MaxRetries:      3,
		Metadata:        make(map[string]any),
		EnableIPv6:      true,
		EnableIPv4:      true,
		RotateServers:   true,
		TTL:             1 * time.Minute,
		NegativeTTL:     30 * time.Second,
		MaxCacheSize:    1000,
		CleanupInterval: 1 * time.Minute,
		Nameservers:     []string{},
		Port:            53,
		MaxStatsSize:    1000,
	}
}

// Validate validates the resolver configuration
func (c *ResolverConfig) Validate() error {
	if c.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	if c.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}
	if c.TTL <= 0 {
		return errors.New("TTL must be positive")
	}
	if c.NegativeTTL <= 0 {
		return errors.New("negative TTL must be positive")
	}
	if c.MaxCacheSize <= 0 {
		return errors.New("max cache size must be positive")
	}
	if c.CleanupInterval <= 0 {
		return errors.New("cleanup interval must be positive")
	}
	if c.MaxStatsSize <= 0 {
		return errors.New("max stats size must be positive")
	}
	if !c.EnableIPv4 && !c.EnableIPv6 {
		return errors.New("at least one of IPv4 or IPv6 must be enabled")
	}
	if c.Port == 0 {
		return errors.New("port cannot be zero")
	}

	// Validate nameserver addresses
	for _, ns := range c.Nameservers {
		if net.ParseIP(ns) == nil {
			return fmt.Errorf("invalid nameserver address: %s", ns)
		}
	}
	return nil
}

// WithTimeout sets the timeout
func (c *ResolverConfig) WithTimeout(d time.Duration) *ResolverConfig {
	c.Timeout = d
	return c
}

// WithMaxRetries sets the maximum number of retries
func (c *ResolverConfig) WithMaxRetries(n int) *ResolverConfig {
	c.MaxRetries = n
	return c
}

// WithTTL sets the TTL
func (c *ResolverConfig) WithTTL(d time.Duration) *ResolverConfig {
	c.TTL = d
	return c
}

// WithNegativeTTL sets the negative TTL
func (c *ResolverConfig) WithNegativeTTL(d time.Duration) *ResolverConfig {
	c.NegativeTTL = d
	return c
}

// WithMaxCacheSize sets the maximum cache size
func (c *ResolverConfig) WithMaxCacheSize(n int) *ResolverConfig {
	c.MaxCacheSize = n
	return c
}

// WithCleanupInterval sets the cleanup interval
func (c *ResolverConfig) WithCleanupInterval(d time.Duration) *ResolverConfig {
	c.CleanupInterval = d
	return c
}

// WithNameservers sets the nameservers
func (c *ResolverConfig) WithNameservers(servers []string) *ResolverConfig {
	c.Nameservers = make([]string, len(servers))
	copy(c.Nameservers, servers)
	return c
}

// WithPort sets the DNS server port
func (c *ResolverConfig) WithPort(port uint16) *ResolverConfig {
	c.Port = port
	return c
}

// WithMetadata sets a metadata value
func (c *ResolverConfig) WithMetadata(key string, value any) *ResolverConfig {
	if c.Metadata == nil {
		c.Metadata = make(map[string]any)
	}
	c.Metadata[key] = value
	return c
}

// WithMaxStatsSize sets the maximum number of lookup times to track
func (c *ResolverConfig) WithMaxStatsSize(n int) *ResolverConfig {
	c.MaxStatsSize = n
	return c
}

// WithIPVersion enables/disables IP versions
func (c *ResolverConfig) WithIPVersion(ipv4, ipv6 bool) *ResolverConfig {
	c.EnableIPv4 = ipv4
	c.EnableIPv6 = ipv6
	return c
}

// WithServerRotation enables/disables server rotation
func (c *ResolverConfig) WithServerRotation(enabled bool) *ResolverConfig {
	c.RotateServers = enabled
	return c
}
