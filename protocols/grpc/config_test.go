package grpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewConfigDefaults(t *testing.T) {
	cfg := NewConfig()
	require.Equal(t, "localhost", cfg.ServerAddress)
	require.Equal(t, 50051, cfg.Port)
	require.Equal(t, 30*time.Second, cfg.Timeout)
	require.Equal(t, 3, cfg.MaxRetries)
	require.Equal(t, uint32(100), cfg.MaxConcurrentStreams)
	require.Equal(t, int32(1024*1024), cfg.InitialWindowSize)
	require.Equal(t, uint32(8192), cfg.MaxHeaderListSize)
	require.Equal(t, 30*time.Second, cfg.KeepAliveTime)
	require.Equal(t, 10*time.Second, cfg.KeepAliveTimeout)
}

func TestConfigValidate(t *testing.T) {
	cfg := NewConfig()
	require.NoError(t, cfg.Validate())
	cfg.ServerAddress = ""
	require.Error(t, cfg.Validate())
	cfg.ServerAddress = "localhost"
	cfg.Port = 0
	require.Error(t, cfg.Validate())
	cfg.Port = 50051
	cfg.Timeout = 0
	require.Error(t, cfg.Validate())
	cfg.Timeout = 1 * time.Second
	cfg.MaxRetries = -1
	require.Error(t, cfg.Validate())
	cfg.MaxRetries = 0
	cfg.MaxConcurrentStreams = 0
	require.Error(t, cfg.Validate())
	cfg.MaxConcurrentStreams = 1
	cfg.InitialWindowSize = 0
	require.Error(t, cfg.Validate())
	cfg.InitialWindowSize = 1
	cfg.MaxHeaderListSize = 0
	require.Error(t, cfg.Validate())
	cfg.MaxHeaderListSize = 1
	cfg.KeepAliveTime = 0
	require.Error(t, cfg.Validate())
	cfg.KeepAliveTime = 1
	cfg.KeepAliveTimeout = 0
	require.Error(t, cfg.Validate())
}

func TestConfigGetFullAddress(t *testing.T) {
	cfg := NewConfig()
	cfg.ServerAddress = "foo"
	cfg.Port = 1234
	require.Equal(t, "foo:1234", cfg.GetFullAddress())
}

func TestConfigWithTLSAndDialOptions(t *testing.T) {
	cfg := NewConfig()
	tls := &TLSConfig{CertFile: "a", KeyFile: "b", CAFile: "c", ServerName: "d", InsecureSkipVerify: true}
	cfg = cfg.WithTLS(tls)
	require.Equal(t, tls, cfg.TLSConfig)
	cfg = cfg.WithDialOptions(nil)
	require.NotNil(t, cfg.DialOptions)
}
