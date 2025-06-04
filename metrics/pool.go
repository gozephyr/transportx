package metrics

import (
	"sync/atomic"
)

// PoolMetricsExporter is a pluggable interface for exporting pool metrics.
// Implementations can use labels for richer observability.
type PoolMetricsExporter interface {
	IncActiveConns(labels map[string]string)
	DecActiveConns(labels map[string]string)
	IncIdleConns(labels map[string]string)
	DecIdleConns(labels map[string]string)
	IncWaitingConns(labels map[string]string)
	DecWaitingConns(labels map[string]string)
	IncTotalConns(labels map[string]string)
	DecTotalConns(labels map[string]string)
	SetMaxActiveConns(n int64, labels map[string]string)
	SetMaxIdleConns(n int64, labels map[string]string)
}

// PoolMetrics holds connection pool statistics.
// Thread-safe for concurrent use.
type PoolMetrics struct {
	ActiveConns    atomic.Int64
	IdleConns      atomic.Int64
	WaitingConns   atomic.Int64
	TotalConns     atomic.Int64
	MaxActiveConns atomic.Int64
	MaxIdleConns   atomic.Int64

	exporter PoolMetricsExporter // Optional exporter for external systems
}

// NewPoolMetrics creates a new PoolMetrics collector.
func NewPoolMetrics() *PoolMetrics {
	return &PoolMetrics{}
}

// SetExporter sets the pool metrics exporter for PoolMetrics.
func (m *PoolMetrics) SetExporter(exporter PoolMetricsExporter) {
	m.exporter = exporter
}

// IncActiveConns increments the active connection count and notifies the exporter.
func (m *PoolMetrics) IncActiveConns(labels map[string]string) {
	m.ActiveConns.Add(1)
	if m.exporter != nil {
		m.exporter.IncActiveConns(labels)
	}
}

// DecActiveConns decrements the active connection count and notifies the exporter.
func (m *PoolMetrics) DecActiveConns(labels map[string]string) {
	m.ActiveConns.Add(-1)
	if m.exporter != nil {
		m.exporter.DecActiveConns(labels)
	}
}

// IncIdleConns increments the idle connection count and notifies the exporter.
func (m *PoolMetrics) IncIdleConns(labels map[string]string) {
	m.IdleConns.Add(1)
	if m.exporter != nil {
		m.exporter.IncIdleConns(labels)
	}
}

// DecIdleConns decrements the idle connection count and notifies the exporter.
func (m *PoolMetrics) DecIdleConns(labels map[string]string) {
	m.IdleConns.Add(-1)
	if m.exporter != nil {
		m.exporter.DecIdleConns(labels)
	}
}

// IncWaitingConns increments the waiting connection count and notifies the exporter.
func (m *PoolMetrics) IncWaitingConns(labels map[string]string) {
	m.WaitingConns.Add(1)
	if m.exporter != nil {
		m.exporter.IncWaitingConns(labels)
	}
}

// DecWaitingConns decrements the waiting connection count and notifies the exporter.
func (m *PoolMetrics) DecWaitingConns(labels map[string]string) {
	m.WaitingConns.Add(-1)
	if m.exporter != nil {
		m.exporter.DecWaitingConns(labels)
	}
}

// IncTotalConns increments the total connection count and notifies the exporter.
func (m *PoolMetrics) IncTotalConns(labels map[string]string) {
	m.TotalConns.Add(1)
	if m.exporter != nil {
		m.exporter.IncTotalConns(labels)
	}
}

// DecTotalConns decrements the total connection count and notifies the exporter.
func (m *PoolMetrics) DecTotalConns(labels map[string]string) {
	m.TotalConns.Add(-1)
	if m.exporter != nil {
		m.exporter.DecTotalConns(labels)
	}
}

// SetMaxActiveConns sets the max active connections and notifies the exporter.
func (m *PoolMetrics) SetMaxActiveConns(n int64, labels map[string]string) {
	m.MaxActiveConns.Store(n)
	if m.exporter != nil {
		m.exporter.SetMaxActiveConns(n, labels)
	}
}

// SetMaxIdleConns sets the max idle connections and notifies the exporter.
func (m *PoolMetrics) SetMaxIdleConns(n int64, labels map[string]string) {
	m.MaxIdleConns.Store(n)
	if m.exporter != nil {
		m.exporter.SetMaxIdleConns(n, labels)
	}
}

// PoolSnapshot is a snapshot of pool metrics.
type PoolSnapshot struct {
	ActiveConns    int64
	IdleConns      int64
	WaitingConns   int64
	TotalConns     int64
	MaxActiveConns int64
	MaxIdleConns   int64
}

// Snapshot returns a snapshot of the current pool metrics.
func (m *PoolMetrics) Snapshot() PoolSnapshot {
	return PoolSnapshot{
		ActiveConns:    m.ActiveConns.Load(),
		IdleConns:      m.IdleConns.Load(),
		WaitingConns:   m.WaitingConns.Load(),
		TotalConns:     m.TotalConns.Load(),
		MaxActiveConns: m.MaxActiveConns.Load(),
		MaxIdleConns:   m.MaxIdleConns.Load(),
	}
}
