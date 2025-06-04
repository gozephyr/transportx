package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// DNSMetricsExporter is a pluggable interface for exporting DNS metrics.
// Implementations can use labels for richer observability.
type DNSMetricsExporter interface {
	IncCacheHits(labels map[string]string)
	IncCacheMisses(labels map[string]string)
	IncLookupErrors(labels map[string]string)
	IncRetryAttempts(labels map[string]string)
	IncCacheErrors(labels map[string]string)
	RecordLookupTime(duration time.Duration, labels map[string]string)
}

// DNSMetrics holds DNS-specific statistics.
// Thread-safe for concurrent use.
type DNSMetrics struct {
	CacheHits     atomic.Int64
	CacheMisses   atomic.Int64
	LookupErrors  atomic.Int64
	RetryAttempts atomic.Int64
	CacheErrors   atomic.Int64
	mu            sync.Mutex
	LookupTimes   []time.Duration
	MaxStatsSize  int

	exporter DNSMetricsExporter // Optional exporter for external systems
}

// NewDNSMetrics creates a new DNSMetrics collector.
func NewDNSMetrics(maxStatsSize int) *DNSMetrics {
	return &DNSMetrics{MaxStatsSize: maxStatsSize}
}

// SetExporter sets the DNS metrics exporter for DNSMetrics.
func (m *DNSMetrics) SetExporter(exporter DNSMetricsExporter) {
	m.exporter = exporter
}

// IncCacheHits increments the cache hit count and notifies the exporter.
func (m *DNSMetrics) IncCacheHits(labels map[string]string) {
	m.CacheHits.Add(1)
	if m.exporter != nil {
		m.exporter.IncCacheHits(labels)
	}
}

// IncCacheMisses increments the cache miss count and notifies the exporter.
func (m *DNSMetrics) IncCacheMisses(labels map[string]string) {
	m.CacheMisses.Add(1)
	if m.exporter != nil {
		m.exporter.IncCacheMisses(labels)
	}
}

// IncLookupErrors increments the lookup error count and notifies the exporter.
func (m *DNSMetrics) IncLookupErrors(labels map[string]string) {
	m.LookupErrors.Add(1)
	if m.exporter != nil {
		m.exporter.IncLookupErrors(labels)
	}
}

// IncRetryAttempts increments the retry attempt count and notifies the exporter.
func (m *DNSMetrics) IncRetryAttempts(labels map[string]string) {
	m.RetryAttempts.Add(1)
	if m.exporter != nil {
		m.exporter.IncRetryAttempts(labels)
	}
}

// IncCacheErrors increments the cache error count and notifies the exporter.
func (m *DNSMetrics) IncCacheErrors(labels map[string]string) {
	m.CacheErrors.Add(1)
	if m.exporter != nil {
		m.exporter.IncCacheErrors(labels)
	}
}

// RecordLookupTime records a lookup time and notifies the exporter.
func (m *DNSMetrics) RecordLookupTime(duration time.Duration, labels map[string]string, isCacheHit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !isCacheHit {
		m.LookupTimes = append(m.LookupTimes, duration)
		if len(m.LookupTimes) > m.MaxStatsSize && m.MaxStatsSize > 0 {
			m.LookupTimes = m.LookupTimes[1:]
		}
		if m.exporter != nil {
			m.exporter.RecordLookupTime(duration, labels)
		}
	}
}

// DNSSnapshot is a snapshot of DNS metrics.
type DNSSnapshot struct {
	CacheHits     int64
	CacheMisses   int64
	LookupErrors  int64
	RetryAttempts int64
	LookupTimes   []time.Duration
}

// Snapshot returns a snapshot of the current DNS metrics.
func (m *DNSMetrics) Snapshot() DNSSnapshot {
	m.mu.Lock()
	lookupTimes := make([]time.Duration, len(m.LookupTimes))
	copy(lookupTimes, m.LookupTimes)
	m.mu.Unlock()
	return DNSSnapshot{
		CacheHits:     m.CacheHits.Load(),
		CacheMisses:   m.CacheMisses.Load(),
		LookupErrors:  m.LookupErrors.Load(),
		RetryAttempts: m.RetryAttempts.Load(),
		LookupTimes:   lookupTimes,
	}
}
