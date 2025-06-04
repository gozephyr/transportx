package dns

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolverConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := DefaultConfig()
		require.NotNil(t, cfg)
		require.Equal(t, 30*time.Second, cfg.Timeout)
		require.Equal(t, 3, cfg.MaxRetries)
		require.NotNil(t, cfg.Metadata)
		require.Equal(t, 1*time.Minute, cfg.TTL)
		require.Equal(t, 1000, cfg.MaxCacheSize)
		require.Equal(t, 1*time.Minute, cfg.CleanupInterval)
		require.Empty(t, cfg.Nameservers)
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		tests := []struct {
			name        string
			modify      func(*ResolverConfig)
			expectError string
		}{
			{
				name: "ValidConfig",
				modify: func(cfg *ResolverConfig) {
					// No modifications needed
				},
				expectError: "",
			},
			{
				name: "InvalidTimeout",
				modify: func(cfg *ResolverConfig) {
					cfg.Timeout = 0
				},
				expectError: "timeout must be positive",
			},
			{
				name: "InvalidMaxRetries",
				modify: func(cfg *ResolverConfig) {
					cfg.MaxRetries = -1
				},
				expectError: "max retries cannot be negative",
			},
			{
				name: "InvalidTTL",
				modify: func(cfg *ResolverConfig) {
					cfg.TTL = 0
				},
				expectError: "TTL must be positive",
			},
			{
				name: "InvalidMaxCacheSize",
				modify: func(cfg *ResolverConfig) {
					cfg.MaxCacheSize = 0
				},
				expectError: "max cache size must be positive",
			},
			{
				name: "InvalidCleanupInterval",
				modify: func(cfg *ResolverConfig) {
					cfg.CleanupInterval = 0
				},
				expectError: "cleanup interval must be positive",
			},
			{
				name: "InvalidMaxStatsSize",
				modify: func(cfg *ResolverConfig) {
					cfg.MaxStatsSize = 0
				},
				expectError: "max stats size must be positive",
			},
			{
				name: "InvalidIPVersion",
				modify: func(cfg *ResolverConfig) {
					cfg.EnableIPv4 = false
					cfg.EnableIPv6 = false
				},
				expectError: "at least one of IPv4 or IPv6 must be enabled",
			},
			{
				name: "InvalidPort",
				modify: func(cfg *ResolverConfig) {
					cfg.Port = 0
				},
				expectError: "port cannot be zero",
			},
			{
				name: "InvalidNameserver",
				modify: func(cfg *ResolverConfig) {
					cfg.Nameservers = []string{"invalid.ip"}
				},
				expectError: "invalid nameserver address",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cfg := DefaultConfig()
				tt.modify(cfg)
				err := cfg.Validate()
				if tt.expectError != "" {
					require.Error(t, err)
					require.Contains(t, err.Error(), tt.expectError)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("ConfigBuilderMethods", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.WithTimeout(10*time.Second).
			WithMaxRetries(5).
			WithTTL(1*time.Minute).
			WithMaxCacheSize(10).
			WithCleanupInterval(2*time.Minute).
			WithNameservers([]string{"8.8.8.8", "1.1.1.1"}).
			WithMetadata("foo", "bar")

		require.Equal(t, 10*time.Second, cfg.Timeout)
		require.Equal(t, 5, cfg.MaxRetries)
		require.Equal(t, 1*time.Minute, cfg.TTL)
		require.Equal(t, 10, cfg.MaxCacheSize)
		require.Equal(t, 2*time.Minute, cfg.CleanupInterval)
		require.Equal(t, []string{"8.8.8.8", "1.1.1.1"}, cfg.Nameservers)
		require.Equal(t, "bar", cfg.Metadata["foo"])
	})

	t.Run("BuilderMethodsCoverage", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.WithPort(5353)
		require.Equal(t, uint16(5353), cfg.Port)
		cfg.WithMaxStatsSize(123)
		require.Equal(t, 123, cfg.MaxStatsSize)
		cfg.WithIPVersion(true, false)
		require.True(t, cfg.EnableIPv4)
		require.False(t, cfg.EnableIPv6)
		cfg.WithServerRotation(false)
		require.False(t, cfg.RotateServers)
		cfg.WithMetadata("key", "value")
		require.Equal(t, "value", cfg.Metadata["key"])
	})
}

func TestResolverCreation(t *testing.T) {
	t.Run("DefaultResolver", func(t *testing.T) {
		r, err := NewResolver(nil)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.NotNil(t, r.cache)
		require.NotNil(t, r.config)
		require.NotNil(t, r.stats)
		require.NotNil(t, r.resolver)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		cfg := DefaultConfig().WithNameservers([]string{"8.8.8.8"})
		r, err := NewResolver(cfg)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.Equal(t, cfg, r.config)
	})
}
