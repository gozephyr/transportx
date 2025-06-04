// Package metrics provides metrics collection for transportx.
//
// Example usage:
//
//	dnsMetrics := metrics.NewDNSMetrics(100)
//	dnsMetrics.SetExporter(myPrometheusExporter) // myPrometheusExporter implements DNSMetricsExporter
//
//	poolMetrics := metrics.NewPoolMetrics()
//	poolMetrics.SetExporter(myPoolExporter) // myPoolExporter implements PoolMetricsExporter
//
// Exporters can use context and labels for richer observability.
//
// To export metrics to Prometheus or another system, implement the MetricsExporter interface
// and register it with SetExporter. The exporter can be implemented in a separate package.
//
// Example:
//
//	type PromExporter struct { /* ... */ }
//	func (e *PromExporter) ExportConnections(value int64, labels map[string]string) { /* ... */ }
//	// ... implement other methods ...
//
//	m := metrics.NewMetrics(nil)
//	m.SetExporter(&PromExporter{})
package metrics

import (
	"sync/atomic"
	"time"
)

// MetricsExporter defines methods for exporting metrics to external systems (e.g., Prometheus).
type MetricsExporter interface {
	ExportConnections(value int64, labels map[string]string)
	ExportActiveStreams(value int64, labels map[string]string)
	ExportFailedConnects(value int64, labels map[string]string)
	ExportLatency(value time.Duration, labels map[string]string)
	ExportThroughput(value int64, labels map[string]string)
	ExportErrors(value int64, labels map[string]string)
	ExportRetries(value int64, labels map[string]string)
	ExportMemoryUsage(value int64, labels map[string]string)
	ExportGoroutines(value int64, labels map[string]string)
}

// Metrics holds performance metrics for a transport.
// Not context-aware; for legacy/transport-level metrics.
type Metrics struct {
	// Configuration
	config *Config

	// Connection metrics
	connections    int64
	activeStreams  int64
	failedConnects int64

	// Performance metrics
	latency    time.Duration
	throughput int64
	errors     int64
	retries    int64

	// Resource metrics
	memoryUsage int64
	goroutines  int64

	exporter MetricsExporter // Pluggable exporter
}

// NewMetrics creates a new metrics collector.
func NewMetrics(cfg *Config) *Metrics {
	if cfg == nil {
		cfg = NewConfig()
	}
	return &Metrics{
		config: cfg,
	}
}

// SetExporter sets the metrics exporter (e.g., Prometheus, custom).
func (m *Metrics) SetExporter(exporter MetricsExporter) {
	m.exporter = exporter
}

// IncrementConnections increments the connection count.
func (m *Metrics) IncrementConnections(labels map[string]string) {
	if m.config.EnableConnections {
		val := atomic.AddInt64(&m.connections, 1)
		if m.exporter != nil {
			m.exporter.ExportConnections(val, labels)
		}
	}
}

// DecrementConnections decrements the connection count.
func (m *Metrics) DecrementConnections(labels map[string]string) {
	if m.config.EnableConnections {
		val := atomic.AddInt64(&m.connections, -1)
		if m.exporter != nil {
			m.exporter.ExportConnections(val, labels)
		}
	}
}

// IncrementActiveStreams increments the active streams count.
func (m *Metrics) IncrementActiveStreams(labels map[string]string) {
	if m.config.EnableActiveStreams {
		val := atomic.AddInt64(&m.activeStreams, 1)
		if m.exporter != nil {
			m.exporter.ExportActiveStreams(val, labels)
		}
	}
}

// DecrementActiveStreams decrements the active streams count.
func (m *Metrics) DecrementActiveStreams(labels map[string]string) {
	if m.config.EnableActiveStreams {
		val := atomic.AddInt64(&m.activeStreams, -1)
		if m.exporter != nil {
			m.exporter.ExportActiveStreams(val, labels)
		}
	}
}

// IncrementFailedConnects increments the failed connections count.
func (m *Metrics) IncrementFailedConnects(labels map[string]string) {
	if m.config.EnableFailedConnects {
		val := atomic.AddInt64(&m.failedConnects, 1)
		if m.exporter != nil {
			m.exporter.ExportFailedConnects(val, labels)
		}
	}
}

// SetLatency sets the current latency.
func (m *Metrics) SetLatency(latency time.Duration, labels map[string]string) {
	if m.config.EnableLatency {
		m.latency = latency
		if m.exporter != nil {
			m.exporter.ExportLatency(latency, labels)
		}
	}
}

// IncrementThroughput increments the throughput count.
func (m *Metrics) IncrementThroughput(labels map[string]string) {
	if m.config.EnableThroughput {
		val := atomic.AddInt64(&m.throughput, 1)
		if m.exporter != nil {
			m.exporter.ExportThroughput(val, labels)
		}
	}
}

// IncrementErrors increments the error count.
func (m *Metrics) IncrementErrors(labels map[string]string) {
	if m.config.EnableErrors {
		val := atomic.AddInt64(&m.errors, 1)
		if m.exporter != nil {
			m.exporter.ExportErrors(val, labels)
		}
	}
}

// IncrementRetries increments the retry count.
func (m *Metrics) IncrementRetries(labels map[string]string) {
	if m.config.EnableRetries {
		val := atomic.AddInt64(&m.retries, 1)
		if m.exporter != nil {
			m.exporter.ExportRetries(val, labels)
		}
	}
}

// SetMemoryUsage sets the current memory usage.
func (m *Metrics) SetMemoryUsage(usage int64, labels map[string]string) {
	if m.config.EnableMemoryUsage {
		atomic.StoreInt64(&m.memoryUsage, usage)
		if m.exporter != nil {
			m.exporter.ExportMemoryUsage(usage, labels)
		}
	}
}

// SetGoroutines sets the current goroutine count.
func (m *Metrics) SetGoroutines(count int64, labels map[string]string) {
	if m.config.EnableGoroutines {
		atomic.StoreInt64(&m.goroutines, count)
		if m.exporter != nil {
			m.exporter.ExportGoroutines(count, labels)
		}
	}
}

// GetConnections returns the current connection count.
func (m *Metrics) GetConnections() int64 {
	return atomic.LoadInt64(&m.connections)
}

// GetActiveStreams returns the current active streams count.
func (m *Metrics) GetActiveStreams() int64 {
	return atomic.LoadInt64(&m.activeStreams)
}

// GetFailedConnects returns the current failed connections count.
func (m *Metrics) GetFailedConnects() int64 {
	return atomic.LoadInt64(&m.failedConnects)
}

// GetLatency returns the current latency.
func (m *Metrics) GetLatency() time.Duration {
	return m.latency
}

// GetThroughput returns the current throughput count.
func (m *Metrics) GetThroughput() int64 {
	return atomic.LoadInt64(&m.throughput)
}

// GetErrors returns the current error count.
func (m *Metrics) GetErrors() int64 {
	return atomic.LoadInt64(&m.errors)
}

// GetRetries returns the current retry count.
func (m *Metrics) GetRetries() int64 {
	return atomic.LoadInt64(&m.retries)
}

// GetMemoryUsage returns the current memory usage.
func (m *Metrics) GetMemoryUsage() int64 {
	return atomic.LoadInt64(&m.memoryUsage)
}

// GetGoroutines returns the current goroutine count.
func (m *Metrics) GetGoroutines() int64 {
	return atomic.LoadInt64(&m.goroutines)
}

// CheckThresholds checks if any metrics exceed their configured thresholds.
func (m *Metrics) CheckThresholds() map[string]bool {
	thresholds := make(map[string]bool)

	thresholds["connections"] = m.checkThreshold(m.GetConnections(), m.config.MaxConnections, m.config.EnableConnections)
	thresholds["active_streams"] = m.checkThreshold(m.GetActiveStreams(), m.config.MaxActiveStreams, m.config.EnableActiveStreams)
	thresholds["failed_connects"] = m.checkThreshold(m.GetFailedConnects(), m.config.MaxFailedConnects, m.config.EnableFailedConnects)
	thresholds["latency"] = m.config.EnableLatency && m.GetLatency() > m.config.MaxLatency
	thresholds["throughput"] = m.checkThreshold(m.GetThroughput(), m.config.MaxThroughput, m.config.EnableThroughput)
	thresholds["errors"] = m.checkThreshold(m.GetErrors(), m.config.MaxErrors, m.config.EnableErrors)
	thresholds["retries"] = m.checkThreshold(m.GetRetries(), m.config.MaxRetries, m.config.EnableRetries)
	thresholds["memory_usage"] = m.checkThreshold(m.GetMemoryUsage(), m.config.MaxMemoryUsage, m.config.EnableMemoryUsage)
	thresholds["goroutines"] = m.checkThreshold(m.GetGoroutines(), m.config.MaxGoroutines, m.config.EnableGoroutines)

	return thresholds
}

// checkThreshold checks if a metric exceeds its threshold.
func (m *Metrics) checkThreshold(value, threshold int64, enabled bool) bool {
	return enabled && value >= threshold
}
