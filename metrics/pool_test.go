package metrics

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockPoolExporter struct {
	calls atomic.Int64
}

func (m *mockPoolExporter) IncActiveConns(labels map[string]string)             { m.calls.Add(1) }
func (m *mockPoolExporter) DecActiveConns(labels map[string]string)             { m.calls.Add(1) }
func (m *mockPoolExporter) IncIdleConns(labels map[string]string)               { m.calls.Add(1) }
func (m *mockPoolExporter) DecIdleConns(labels map[string]string)               { m.calls.Add(1) }
func (m *mockPoolExporter) IncWaitingConns(labels map[string]string)            { m.calls.Add(1) }
func (m *mockPoolExporter) DecWaitingConns(labels map[string]string)            { m.calls.Add(1) }
func (m *mockPoolExporter) IncTotalConns(labels map[string]string)              { m.calls.Add(1) }
func (m *mockPoolExporter) DecTotalConns(labels map[string]string)              { m.calls.Add(1) }
func (m *mockPoolExporter) SetMaxActiveConns(n int64, labels map[string]string) { m.calls.Add(1) }
func (m *mockPoolExporter) SetMaxIdleConns(n int64, labels map[string]string)   { m.calls.Add(1) }

func TestPoolMetricsBasic(t *testing.T) {
	m := NewPoolMetrics()
	labels := map[string]string{"pool": "test"}
	m.IncActiveConns(labels)
	m.IncIdleConns(labels)
	m.IncWaitingConns(labels)
	m.IncTotalConns(labels)
	m.SetMaxActiveConns(10, labels)
	m.SetMaxIdleConns(5, labels)
	require.Equal(t, int64(1), m.ActiveConns.Load())
	require.Equal(t, int64(1), m.IdleConns.Load())
	require.Equal(t, int64(1), m.WaitingConns.Load())
	require.Equal(t, int64(1), m.TotalConns.Load())
	require.Equal(t, int64(10), m.MaxActiveConns.Load())
	require.Equal(t, int64(5), m.MaxIdleConns.Load())
	m.DecActiveConns(labels)
	m.DecIdleConns(labels)
	m.DecWaitingConns(labels)
	m.DecTotalConns(labels)
	require.Equal(t, int64(0), m.ActiveConns.Load())
	require.Equal(t, int64(0), m.IdleConns.Load())
	require.Equal(t, int64(0), m.WaitingConns.Load())
	require.Equal(t, int64(0), m.TotalConns.Load())
}

func TestPoolMetricsExporter(t *testing.T) {
	m := NewPoolMetrics()
	exp := &mockPoolExporter{}
	m.SetExporter(exp)
	labels := map[string]string{"pool": "exp"}
	m.IncActiveConns(labels)
	m.DecActiveConns(labels)
	m.IncIdleConns(labels)
	m.DecIdleConns(labels)
	m.IncWaitingConns(labels)
	m.DecWaitingConns(labels)
	m.IncTotalConns(labels)
	m.DecTotalConns(labels)
	m.SetMaxActiveConns(10, labels)
	m.SetMaxIdleConns(5, labels)
	require.GreaterOrEqual(t, exp.calls.Load(), int64(10))
}

func TestPoolMetricsSnapshot(t *testing.T) {
	m := NewPoolMetrics()
	labels := map[string]string{"pool": "snap"}
	m.IncActiveConns(labels)
	m.IncIdleConns(labels)
	m.IncWaitingConns(labels)
	m.IncTotalConns(labels)
	m.SetMaxActiveConns(10, labels)
	m.SetMaxIdleConns(5, labels)
	snap := m.Snapshot()
	require.Equal(t, int64(1), snap.ActiveConns)
	require.Equal(t, int64(1), snap.IdleConns)
	require.Equal(t, int64(1), snap.WaitingConns)
	require.Equal(t, int64(1), snap.TotalConns)
	require.Equal(t, int64(10), snap.MaxActiveConns)
	require.Equal(t, int64(5), snap.MaxIdleConns)
}

func TestPoolMetricsConcurrencyAndEdgeCases(t *testing.T) {
	m := NewPoolMetrics()
	labels := map[string]string{"pool": "concurrent"}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				m.IncActiveConns(labels)
				m.DecActiveConns(labels)
				m.IncIdleConns(labels)
				m.DecIdleConns(labels)
				m.IncWaitingConns(labels)
				m.DecWaitingConns(labels)
				m.IncTotalConns(labels)
				m.DecTotalConns(labels)
				m.SetMaxActiveConns(int64(j), labels)
				m.SetMaxIdleConns(int64(j), labels)
			}
		}()
	}
	wg.Wait()
	// Should be 0 after equal inc/dec
	require.Equal(t, int64(0), m.ActiveConns.Load())
	require.Equal(t, int64(0), m.IdleConns.Load())
	require.Equal(t, int64(0), m.WaitingConns.Load())
	require.Equal(t, int64(0), m.TotalConns.Load())
	// Max values should be last set
	require.Equal(t, int64(999), m.MaxActiveConns.Load())
	require.Equal(t, int64(999), m.MaxIdleConns.Load())
}

func TestPoolMetricsNoExporter(t *testing.T) {
	m := NewPoolMetrics()
	labels := map[string]string{"pool": "noexp"}
	// Should not panic or error
	m.IncActiveConns(labels)
	m.DecActiveConns(labels)
	m.IncIdleConns(labels)
	m.DecIdleConns(labels)
	m.IncWaitingConns(labels)
	m.DecWaitingConns(labels)
	m.IncTotalConns(labels)
	m.DecTotalConns(labels)
	m.SetMaxActiveConns(-1, labels)
	m.SetMaxIdleConns(-1, labels)
	require.Equal(t, int64(-1), m.MaxActiveConns.Load())
	require.Equal(t, int64(-1), m.MaxIdleConns.Load())
}
