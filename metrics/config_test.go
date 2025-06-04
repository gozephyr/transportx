package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfigDefaults(t *testing.T) {
	cfg := NewConfig()
	require.True(t, cfg.EnableConnections)
	require.True(t, cfg.EnableActiveStreams)
	require.True(t, cfg.EnableFailedConnects)
	require.True(t, cfg.EnableLatency)
	require.True(t, cfg.EnableThroughput)
	require.True(t, cfg.EnableErrors)
	require.True(t, cfg.EnableRetries)
	require.True(t, cfg.EnableMemoryUsage)
	require.True(t, cfg.EnableGoroutines)
	require.Equal(t, int64(1000), cfg.MaxConnections)
	require.Equal(t, int64(100), cfg.MaxActiveStreams)
	require.Equal(t, int64(100), cfg.MaxFailedConnects)
	require.Equal(t, 5*time.Second, cfg.MaxLatency)
	require.Equal(t, int64(10000), cfg.MaxThroughput)
	require.Equal(t, int64(100), cfg.MaxErrors)
	require.Equal(t, int64(10), cfg.MaxRetries)
	require.Equal(t, int64(1024*1024*1024), cfg.MaxMemoryUsage)
	require.Equal(t, int64(1000), cfg.MaxGoroutines)
}

func TestConfigWithers(t *testing.T) {
	cfg := NewConfig().WithConnections(false).WithActiveStreams(false).WithFailedConnects(false).WithLatency(false).WithThroughput(false).WithErrors(false).WithRetries(false).WithMemoryUsage(false).WithGoroutines(false).WithMaxConnections(1).WithMaxActiveStreams(2).WithMaxFailedConnects(3).WithMaxLatency(4 * time.Second).WithMaxThroughput(5).WithMaxErrors(6).WithMaxRetries(7).WithMaxMemoryUsage(8).WithMaxGoroutines(9)
	require.False(t, cfg.EnableConnections)
	require.False(t, cfg.EnableActiveStreams)
	require.False(t, cfg.EnableFailedConnects)
	require.False(t, cfg.EnableLatency)
	require.False(t, cfg.EnableThroughput)
	require.False(t, cfg.EnableErrors)
	require.False(t, cfg.EnableRetries)
	require.False(t, cfg.EnableMemoryUsage)
	require.False(t, cfg.EnableGoroutines)
	require.Equal(t, int64(1), cfg.MaxConnections)
	require.Equal(t, int64(2), cfg.MaxActiveStreams)
	require.Equal(t, int64(3), cfg.MaxFailedConnects)
	require.Equal(t, 4*time.Second, cfg.MaxLatency)
	require.Equal(t, int64(5), cfg.MaxThroughput)
	require.Equal(t, int64(6), cfg.MaxErrors)
	require.Equal(t, int64(7), cfg.MaxRetries)
	require.Equal(t, int64(8), cfg.MaxMemoryUsage)
	require.Equal(t, int64(9), cfg.MaxGoroutines)
}

func TestConfigWithersPointerIdentity(t *testing.T) {
	cfg := NewConfig()
	require.Same(t, cfg, cfg.WithConnections(true))
	require.Same(t, cfg, cfg.WithActiveStreams(true))
	require.Same(t, cfg, cfg.WithFailedConnects(true))
	require.Same(t, cfg, cfg.WithLatency(true))
	require.Same(t, cfg, cfg.WithThroughput(true))
	require.Same(t, cfg, cfg.WithErrors(true))
	require.Same(t, cfg, cfg.WithRetries(true))
	require.Same(t, cfg, cfg.WithMemoryUsage(true))
	require.Same(t, cfg, cfg.WithGoroutines(true))
	require.Same(t, cfg, cfg.WithMaxConnections(1))
	require.Same(t, cfg, cfg.WithMaxActiveStreams(1))
	require.Same(t, cfg, cfg.WithMaxFailedConnects(1))
	require.Same(t, cfg, cfg.WithMaxLatency(1*time.Second))
	require.Same(t, cfg, cfg.WithMaxThroughput(1))
	require.Same(t, cfg, cfg.WithMaxErrors(1))
	require.Same(t, cfg, cfg.WithMaxRetries(1))
	require.Same(t, cfg, cfg.WithMaxMemoryUsage(1))
	require.Same(t, cfg, cfg.WithMaxGoroutines(1))
}

func TestConfigThresholdsEdgeCases(t *testing.T) {
	cfg := NewConfig().WithMaxConnections(0).WithMaxActiveStreams(-1).WithMaxFailedConnects(-2).WithMaxLatency(0).WithMaxThroughput(-3).WithMaxErrors(-4).WithMaxRetries(-5).WithMaxMemoryUsage(-6).WithMaxGoroutines(-7)
	require.Equal(t, int64(0), cfg.MaxConnections)
	require.Equal(t, int64(-1), cfg.MaxActiveStreams)
	require.Equal(t, int64(-2), cfg.MaxFailedConnects)
	require.Equal(t, time.Duration(0), cfg.MaxLatency)
	require.Equal(t, int64(-3), cfg.MaxThroughput)
	require.Equal(t, int64(-4), cfg.MaxErrors)
	require.Equal(t, int64(-5), cfg.MaxRetries)
	require.Equal(t, int64(-6), cfg.MaxMemoryUsage)
	require.Equal(t, int64(-7), cfg.MaxGoroutines)
}
