package metrics

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockDNSExporter struct {
	calls atomic.Int64
}

func (m *mockDNSExporter) IncCacheHits(labels map[string]string)     { m.calls.Add(1) }
func (m *mockDNSExporter) IncCacheMisses(labels map[string]string)   { m.calls.Add(1) }
func (m *mockDNSExporter) IncLookupErrors(labels map[string]string)  { m.calls.Add(1) }
func (m *mockDNSExporter) IncRetryAttempts(labels map[string]string) { m.calls.Add(1) }
func (m *mockDNSExporter) IncCacheErrors(labels map[string]string)   { m.calls.Add(1) }
func (m *mockDNSExporter) RecordLookupTime(duration time.Duration, labels map[string]string) {
	m.calls.Add(1)
}

func TestDNSMetricsBasicAndSnapshot(t *testing.T) {
	m := NewDNSMetrics(10)
	labels := map[string]string{"dns": "test"}
	m.IncCacheHits(labels)
	m.IncCacheMisses(labels)
	m.IncLookupErrors(labels)
	m.IncRetryAttempts(labels)
	m.IncCacheErrors(labels)
	m.RecordLookupTime(100*time.Millisecond, labels, false)
	snap := m.Snapshot()
	require.Equal(t, int64(1), snap.CacheHits)
	require.Equal(t, int64(1), snap.CacheMisses)
	require.Equal(t, int64(1), snap.LookupErrors)
	require.Equal(t, int64(1), snap.RetryAttempts)
	require.Len(t, snap.LookupTimes, 1)
}

func TestDNSMetricsExporterIntegration(t *testing.T) {
	m := NewDNSMetrics(10)
	exp := &mockDNSExporter{}
	m.SetExporter(exp)
	labels := map[string]string{"dns": "exp"}
	m.IncCacheHits(labels)
	m.IncCacheMisses(labels)
	m.IncLookupErrors(labels)
	m.IncRetryAttempts(labels)
	m.IncCacheErrors(labels)
	m.RecordLookupTime(100*time.Millisecond, labels, false)
	require.GreaterOrEqual(t, exp.calls.Load(), int64(6))
}

func TestDNSMetricsLookupTimeCap(t *testing.T) {
	tests := []struct {
		name        string
		maxStats    int
		expectedLen int
	}{
		{"Cap3", 3, 3},
		{"ZeroCap", 0, 10},
		{"NegativeCap", -1, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewDNSMetrics(tt.maxStats)
			labels := map[string]string{"dns": tt.name}
			for i := 0; i < 10; i++ {
				m.RecordLookupTime(time.Duration(i)*time.Millisecond, labels, false)
			}
			snap := m.Snapshot()
			require.Len(t, snap.LookupTimes, tt.expectedLen)
			if tt.expectedLen == 3 {
				require.Equal(t, 7*time.Millisecond, snap.LookupTimes[0])
				require.Equal(t, 8*time.Millisecond, snap.LookupTimes[1])
				require.Equal(t, 9*time.Millisecond, snap.LookupTimes[2])
			}
		})
	}
}

func TestDNSMetricsConcurrency(t *testing.T) {
	m := NewDNSMetrics(5)
	labels := map[string]string{"dns": "concurrent"}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.IncCacheHits(labels)
				m.IncCacheMisses(labels)
				m.IncLookupErrors(labels)
				m.IncRetryAttempts(labels)
				m.IncCacheErrors(labels)
				m.RecordLookupTime(time.Duration(j)*time.Millisecond, labels, false)
			}
		}()
	}
	wg.Wait()
	snap := m.Snapshot()
	require.GreaterOrEqual(t, snap.CacheHits, int64(2000))
	require.GreaterOrEqual(t, snap.CacheMisses, int64(2000))
	require.GreaterOrEqual(t, snap.LookupErrors, int64(2000))
	require.GreaterOrEqual(t, snap.RetryAttempts, int64(2000))
	require.GreaterOrEqual(t, snap.LookupTimes[0], time.Duration(95)*time.Millisecond)
	require.Len(t, snap.LookupTimes, 5)
}

func TestDNSMetricsNoExporter(t *testing.T) {
	m := NewDNSMetrics(2)
	labels := map[string]string{"dns": "noexp"}
	m.IncCacheHits(labels)
	m.IncCacheMisses(labels)
	m.IncLookupErrors(labels)
	m.IncRetryAttempts(labels)
	m.IncCacheErrors(labels)
	m.RecordLookupTime(1*time.Millisecond, labels, false)
	m.RecordLookupTime(2*time.Millisecond, labels, false)
	m.RecordLookupTime(3*time.Millisecond, labels, false)
	snap := m.Snapshot()
	require.Len(t, snap.LookupTimes, 2)
	require.Equal(t, 2*time.Millisecond, snap.LookupTimes[0])
	require.Equal(t, 3*time.Millisecond, snap.LookupTimes[1])
}
