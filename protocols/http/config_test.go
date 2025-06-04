package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewConfigDefaults(t *testing.T) {
	cfg := NewConfig()
	require.Equal(t, "http", cfg.Protocol)
	require.Equal(t, "localhost", cfg.ServerAddress)
	require.Equal(t, 8080, cfg.Port)
	require.Equal(t, 100, cfg.MaxIdleConns)
	require.Equal(t, 100, cfg.MaxConnsPerHost)
	require.Equal(t, 90*time.Second, cfg.IdleTimeout)
	require.Equal(t, 30*time.Second, cfg.KeepAlive)
	require.Equal(t, 30*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 5*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 30*time.Second, cfg.HealthCheckDelay)
	require.True(t, cfg.EnableHTTP2)
	require.Equal(t, 32*1024, cfg.MaxHeaderBytes)
	require.False(t, cfg.DisableCompression)
	require.False(t, cfg.DisableKeepAlives)
}

func TestConfigGetFullURL(t *testing.T) {
	cfg := NewConfig()
	cfg.ServerAddress = "example.com"
	cfg.Port = 8080
	require.Equal(t, "http://example.com:8080", cfg.GetFullURL())
	cfg.Port = 0
	require.Equal(t, "http://example.com", cfg.GetFullURL())
}

func TestConfigValidateError(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*Config)
		errorS string
	}{
		{"ValidConfig", func(c *Config) {}, ""},
		{"EmptyServerAddress", func(c *Config) { c.ServerAddress = "" }, "server address is required"},
		{"InvalidServerAddress", func(c *Config) { c.ServerAddress = ":bad" }, "invalid server address"},
		{"MaxIdleConnsZero", func(c *Config) { c.MaxIdleConns = 0 }, "max idle connections must be positive"},
		{"MaxConnsPerHostZero", func(c *Config) { c.MaxConnsPerHost = 0 }, "max connections per host must be positive"},
		{"IdleTimeoutZero", func(c *Config) { c.IdleTimeout = 0 }, "idle timeout must be positive"},
		{"KeepAliveZero", func(c *Config) { c.KeepAlive = 0 }, "keepalive must be positive"},
		{"ResponseTimeoutZero", func(c *Config) { c.ResponseTimeout = 0 }, "response timeout must be positive"},
		{"MaxWaitDurationZero", func(c *Config) { c.MaxWaitDuration = 0 }, "max wait duration must be positive"},
		{"HealthCheckDelayZero", func(c *Config) { c.HealthCheckDelay = 0 }, "health check delay must be positive"},
		{"MaxHeaderBytesZero", func(c *Config) { c.MaxHeaderBytes = 0 }, "max header bytes must be between 1 and"},
		{"MaxHeaderBytesTooLarge", func(c *Config) { c.MaxHeaderBytes = 2 * 1024 * 1024 }, "max header bytes must be between 1 and"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			tt.setup(cfg)
			err := cfg.Validate()
			if tt.errorS == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorS)
			}
		})
	}
}

func TestConfigClone(t *testing.T) {
	cfg1 := NewConfig()
	cfg1.ServerAddress = "a.com"
	cfg1.MaxIdleConns = 10
	cfg1.MaxConnsPerHost = 20
	cfg1.IdleTimeout = 1 * time.Second
	cfg1.KeepAlive = 2 * time.Second
	cfg1.ResponseTimeout = 3 * time.Second
	cfg1.MaxWaitDuration = 4 * time.Second
	cfg1.HealthCheckDelay = 5 * time.Second
	cfg1.MaxHeaderBytes = 4096
	cfg1.EnableHTTP2 = false
	cfg1.DisableCompression = true
	cfg1.DisableKeepAlives = true

	cfg2 := NewConfig()
	cfg2.ServerAddress = "b.com"
	cfg2.MaxIdleConns = 11
	cfg2.MaxConnsPerHost = 21
	cfg2.IdleTimeout = 11 * time.Second
	cfg2.KeepAlive = 12 * time.Second
	cfg2.ResponseTimeout = 13 * time.Second
	cfg2.MaxWaitDuration = 14 * time.Second
	cfg2.HealthCheckDelay = 15 * time.Second
	cfg2.MaxHeaderBytes = 8192
	cfg2.EnableHTTP2 = true
	cfg2.DisableCompression = false
	cfg2.DisableKeepAlives = false

	merged, err := cfg1.Merge(cfg2)
	require.NoError(t, err)
	m := merged.(*Config)
	require.Equal(t, "b.com", m.ServerAddress)
	require.Equal(t, 11, m.MaxIdleConns)
	require.Equal(t, 21, m.MaxConnsPerHost)
	require.Equal(t, 11*time.Second, m.IdleTimeout)
	require.Equal(t, 12*time.Second, m.KeepAlive)
	require.Equal(t, 13*time.Second, m.ResponseTimeout)
	require.Equal(t, 14*time.Second, m.MaxWaitDuration)
	require.Equal(t, 15*time.Second, m.HealthCheckDelay)
	require.Equal(t, 8192, m.MaxHeaderBytes)
	require.True(t, m.EnableHTTP2)
	require.False(t, m.DisableCompression)
	require.False(t, m.DisableKeepAlives)

	clone := cfg1.Clone().(*Config)
	require.Equal(t, cfg1.ServerAddress, clone.ServerAddress)
	require.Equal(t, cfg1.MaxIdleConns, clone.MaxIdleConns)
	require.Equal(t, cfg1.MaxConnsPerHost, clone.MaxConnsPerHost)
	require.Equal(t, cfg1.IdleTimeout, clone.IdleTimeout)
	require.Equal(t, cfg1.KeepAlive, clone.KeepAlive)
	require.Equal(t, cfg1.ResponseTimeout, clone.ResponseTimeout)
	require.Equal(t, cfg1.MaxWaitDuration, clone.MaxWaitDuration)
	require.Equal(t, cfg1.HealthCheckDelay, clone.HealthCheckDelay)
	require.Equal(t, cfg1.MaxHeaderBytes, clone.MaxHeaderBytes)
	require.Equal(t, cfg1.EnableHTTP2, clone.EnableHTTP2)
	require.Equal(t, cfg1.DisableCompression, clone.DisableCompression)
	require.Equal(t, cfg1.DisableKeepAlives, clone.DisableKeepAlives)
}

func TestConfigWithMethods(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithTimeouts(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithLimits(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithAddress(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithTLS(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithHeaders(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithProxy(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithCustomDialer(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithCustomTransport(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithCustomClient(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithCustomResolver(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithCustomPool(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithCustomBuffer(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithCustomLogger(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithCustomMetrics(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithCustomOptions(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithAllCustom(t *testing.T) {
	cfg := NewConfig().
		WithTimeout(1*time.Second).
		WithMaxRetries(2).
		WithServerAddress("foo.com").
		WithMaxIdleConns(3).
		WithMaxConnsPerHost(4).
		WithIdleTimeout(5*time.Second).
		WithKeepAlive(6*time.Second).
		WithResponseTimeout(7*time.Second).
		WithMaxWaitDuration(8*time.Second).
		WithHealthCheckDelay(9*time.Second).
		WithHTTP2(false).
		WithMaxHeaderBytes(10).
		WithCompression(false).
		WithKeepAlives(false).
		WithMetadata("foo", "bar")

	require.Equal(t, 1*time.Second, cfg.Timeout)
	require.Equal(t, 2, cfg.MaxRetries)
	require.Equal(t, "foo.com", cfg.ServerAddress)
	require.Equal(t, 3, cfg.MaxIdleConns)
	require.Equal(t, 4, cfg.MaxConnsPerHost)
	require.Equal(t, 5*time.Second, cfg.IdleTimeout)
	require.Equal(t, 6*time.Second, cfg.KeepAlive)
	require.Equal(t, 7*time.Second, cfg.ResponseTimeout)
	require.Equal(t, 8*time.Second, cfg.MaxWaitDuration)
	require.Equal(t, 9*time.Second, cfg.HealthCheckDelay)
	require.False(t, cfg.EnableHTTP2)
	require.Equal(t, 10, cfg.MaxHeaderBytes)
	require.True(t, cfg.DisableCompression)
	require.True(t, cfg.DisableKeepAlives)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}

func TestConfigWithInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*Config)
		errorS string
	}{
		{"ValidConfig", func(c *Config) {}, ""},
		{"EmptyServerAddress", func(c *Config) { c.ServerAddress = "" }, "server address is required"},
		{"InvalidServerAddress", func(c *Config) { c.ServerAddress = ":bad" }, "invalid server address"},
		{"MaxIdleConnsZero", func(c *Config) { c.MaxIdleConns = 0 }, "max idle connections must be positive"},
		{"MaxConnsPerHostZero", func(c *Config) { c.MaxConnsPerHost = 0 }, "max connections per host must be positive"},
		{"IdleTimeoutZero", func(c *Config) { c.IdleTimeout = 0 }, "idle timeout must be positive"},
		{"KeepAliveZero", func(c *Config) { c.KeepAlive = 0 }, "keepalive must be positive"},
		{"ResponseTimeoutZero", func(c *Config) { c.ResponseTimeout = 0 }, "response timeout must be positive"},
		{"MaxWaitDurationZero", func(c *Config) { c.MaxWaitDuration = 0 }, "max wait duration must be positive"},
		{"HealthCheckDelayZero", func(c *Config) { c.HealthCheckDelay = 0 }, "health check delay must be positive"},
		{"MaxHeaderBytesZero", func(c *Config) { c.MaxHeaderBytes = 0 }, "max header bytes must be between 1 and"},
		{"MaxHeaderBytesTooLarge", func(c *Config) { c.MaxHeaderBytes = 2 * 1024 * 1024 }, "max header bytes must be between 1 and"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			tt.setup(cfg)
			err := cfg.Validate()
			if tt.errorS == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorS)
			}
		})
	}
}

func TestSetMetadata(t *testing.T) {
	cfg := NewConfig()
	cfg.SetMetadata("k", "v")
	require.Equal(t, "v", cfg.Metadata["k"])
}
