package protocols

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewProtocolConfigDefaults(t *testing.T) {
	cfg := NewProtocolConfig()
	require.Equal(t, 30*time.Second, cfg.Timeout)
	require.Equal(t, 3, cfg.MaxRetries)
	require.Equal(t, 32*1024, cfg.BufferSize)
	require.False(t, cfg.EnableTLS)
	require.Equal(t, 30*time.Second, cfg.KeepAlive)
	require.Equal(t, 100, cfg.MaxConnections)
	require.Equal(t, 60*time.Second, cfg.IdleTimeout)
	require.Equal(t, 1*time.Second, cfg.ReconnectDelay)
	require.Equal(t, 5, cfg.MaxReconnectAttempts)
	require.NotNil(t, cfg.Metadata)
}

func TestProtocolConfigValidate(t *testing.T) {
	cfg := NewProtocolConfig()
	require.NoError(t, cfg.Validate())

	cfg.Timeout = 0
	require.Error(t, cfg.Validate())
	cfg.Timeout = 1 * time.Second
	cfg.MaxRetries = -1
	require.Error(t, cfg.Validate())
	cfg.MaxRetries = 1
	cfg.BufferSize = 0
	require.Error(t, cfg.Validate())
	cfg.BufferSize = 1
	cfg.MaxConnections = 0
	require.Error(t, cfg.Validate())
	cfg.MaxConnections = 1
	cfg.IdleTimeout = 0
	require.Error(t, cfg.Validate())
	cfg.IdleTimeout = 1
	cfg.ReconnectDelay = 0
	require.Error(t, cfg.Validate())
	cfg.ReconnectDelay = 1
	cfg.MaxReconnectAttempts = -1
	require.Error(t, cfg.Validate())
}

func TestProtocolConfigWithMethods(t *testing.T) {
	cfg := NewProtocolConfig().
		WithTimeout(2*time.Second).
		WithMaxRetries(4).
		WithBufferSize(2048).
		WithTLS("cert", "key").
		WithKeepAlive(3*time.Second).
		WithMaxConnections(10).
		WithIdleTimeout(4*time.Second).
		WithReconnectDelay(5*time.Second).
		WithMaxReconnectAttempts(7).
		WithMetadata("foo", "bar")

	require.Equal(t, 2*time.Second, cfg.Timeout)
	require.Equal(t, 4, cfg.MaxRetries)
	require.Equal(t, 2048, cfg.BufferSize)
	require.True(t, cfg.EnableTLS)
	require.Equal(t, "cert", cfg.TLSCertFile)
	require.Equal(t, "key", cfg.TLSKeyFile)
	require.Equal(t, 3*time.Second, cfg.KeepAlive)
	require.Equal(t, 10, cfg.MaxConnections)
	require.Equal(t, 4*time.Second, cfg.IdleTimeout)
	require.Equal(t, 5*time.Second, cfg.ReconnectDelay)
	require.Equal(t, 7, cfg.MaxReconnectAttempts)
	require.Equal(t, "bar", cfg.Metadata["foo"])
}
