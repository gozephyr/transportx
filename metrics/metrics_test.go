package metrics

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockExporter struct {
	calls             atomic.Int64
	lastLabels        map[string]string
	lastValueInt      int64
	lastValueDuration time.Duration
}

func (m *mockExporter) ExportConnections(value int64, labels map[string]string) {
	m.calls.Add(1)
	m.lastLabels = labels
	m.lastValueInt = value
}

func (m *mockExporter) ExportActiveStreams(value int64, labels map[string]string) {
	m.calls.Add(1)
	m.lastLabels = labels
	m.lastValueInt = value
}

func (m *mockExporter) ExportFailedConnects(value int64, labels map[string]string) {
	m.calls.Add(1)
	m.lastLabels = labels
	m.lastValueInt = value
}

func (m *mockExporter) ExportLatency(value time.Duration, labels map[string]string) {
	m.calls.Add(1)
	m.lastLabels = labels
	m.lastValueDuration = value
}

func (m *mockExporter) ExportThroughput(value int64, labels map[string]string) {
	m.calls.Add(1)
	m.lastLabels = labels
	m.lastValueInt = value
}

func (m *mockExporter) ExportErrors(value int64, labels map[string]string) {
	m.calls.Add(1)
	m.lastLabels = labels
	m.lastValueInt = value
}

func (m *mockExporter) ExportRetries(value int64, labels map[string]string) {
	m.calls.Add(1)
	m.lastLabels = labels
	m.lastValueInt = value
}

func (m *mockExporter) ExportMemoryUsage(value int64, labels map[string]string) {
	m.calls.Add(1)
	m.lastLabels = labels
	m.lastValueInt = value
}

func (m *mockExporter) ExportGoroutines(value int64, labels map[string]string) {
	m.calls.Add(1)
	m.lastLabels = labels
	m.lastValueInt = value
}

func TestMetricsBasic(t *testing.T) {
	cfg := NewConfig()
	m := NewMetrics(cfg)
	labels := map[string]string{"foo": "bar"}

	m.IncrementConnections(labels)
	require.Equal(t, int64(1), m.GetConnections())
	m.DecrementConnections(labels)
	require.Equal(t, int64(0), m.GetConnections())

	m.IncrementActiveStreams(labels)
	require.Equal(t, int64(1), m.GetActiveStreams())
	m.DecrementActiveStreams(labels)
	require.Equal(t, int64(0), m.GetActiveStreams())

	m.IncrementFailedConnects(labels)
	require.Equal(t, int64(1), m.GetFailedConnects())

	m.SetLatency(123*time.Millisecond, labels)
	require.Equal(t, 123*time.Millisecond, m.GetLatency())

	m.IncrementThroughput(labels)
	require.Equal(t, int64(1), m.GetThroughput())

	m.IncrementErrors(labels)
	require.Equal(t, int64(1), m.GetErrors())

	m.IncrementRetries(labels)
	require.Equal(t, int64(1), m.GetRetries())

	m.SetMemoryUsage(42, labels)
	require.Equal(t, int64(42), m.GetMemoryUsage())

	m.SetGoroutines(7, labels)
	require.Equal(t, int64(7), m.GetGoroutines())
}

func TestMetricsExporter(t *testing.T) {
	cfg := NewConfig()
	m := NewMetrics(cfg)
	exp := &mockExporter{}
	m.SetExporter(exp)
	labels := map[string]string{"x": "y"}

	m.IncrementConnections(labels)
	require.Greater(t, exp.calls.Load(), int64(0))
	m.DecrementConnections(labels)
	m.IncrementActiveStreams(labels)
	m.DecrementActiveStreams(labels)
	m.IncrementFailedConnects(labels)
	m.SetLatency(1*time.Second, labels)
	m.IncrementThroughput(labels)
	m.IncrementErrors(labels)
	m.IncrementRetries(labels)
	m.SetMemoryUsage(100, labels)
	m.SetGoroutines(5, labels)
	require.GreaterOrEqual(t, exp.calls.Load(), int64(10))
}

func TestMetricsThresholds(t *testing.T) {
	cfg := NewConfig().WithMaxConnections(1).WithMaxActiveStreams(1).WithMaxFailedConnects(1).WithMaxLatency(1 * time.Millisecond).WithMaxThroughput(1).WithMaxErrors(1).WithMaxRetries(1).WithMaxMemoryUsage(1).WithMaxGoroutines(1)
	m := NewMetrics(cfg)
	labels := map[string]string{}

	m.IncrementConnections(labels)
	m.IncrementActiveStreams(labels)
	m.IncrementFailedConnects(labels)
	m.SetLatency(2*time.Millisecond, labels)
	m.IncrementThroughput(labels)
	m.IncrementErrors(labels)
	m.IncrementRetries(labels)
	m.SetMemoryUsage(2, labels)
	m.SetGoroutines(2, labels)

	thresh := m.CheckThresholds()
	for k, v := range thresh {
		require.True(t, v, "expected threshold for %s to be true", k)
	}
}

func TestMetricsConcurrency(t *testing.T) {
	cfg := NewConfig()
	m := NewMetrics(cfg)
	labels := map[string]string{"thread": "safe"}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				m.IncrementConnections(labels)
				m.IncrementActiveStreams(labels)
				m.IncrementFailedConnects(labels)
				m.IncrementThroughput(labels)
				m.IncrementErrors(labels)
				m.IncrementRetries(labels)
				m.SetMemoryUsage(int64(j), labels)
				m.SetGoroutines(int64(j), labels)
			}
		}()
	}
	wg.Wait()
	require.Equal(t, int64(100*1000), m.GetConnections())
	require.Equal(t, int64(100*1000), m.GetActiveStreams())
}
