// Package dns provides a robust DNS resolution functionality with caching,
// concurrent lookups, and comprehensive statistics tracking.
package dns

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gozephyr/gencache"
	"github.com/gozephyr/gencache/policy"
	"github.com/gozephyr/gencache/ttl"
	"github.com/gozephyr/transportx/metrics"
)

// Resolver represents a robust DNS resolver
type Resolver struct {
	cache        gencache.Cache[string, *cacheEntry]
	config       *ResolverConfig
	resolver     *net.Resolver
	stats        *metrics.DNSMetrics
	lookupIPFunc func(ctx context.Context, network, host string) ([]net.IP, error) // for testing
	serverIndex  atomic.Uint32                                                     // for server rotation
	mu           sync.RWMutex                                                      // protects resolver modifications
}

type cacheEntry struct {
	IPs        []net.IP
	ExpiresAt  time.Time
	IsNegative bool // Marks negative cache entries
}

// NewResolver creates a new DNS resolver with the given configuration
func NewResolver(cfg *ResolverConfig) (*Resolver, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid resolver configuration: %w", err)
	}

	// Create gencache with TTL and LRU policy
	var ttlConfig ttl.Config
	if cfg.TTL < time.Second {
		// Use test config for very short TTLs (testing scenarios)
		ttlConfig = ttl.TestConfig()
	} else {
		ttlConfig = ttl.DefaultConfig()
	}

	cache := gencache.New(
		gencache.WithTTLConfig[string, *cacheEntry](ttlConfig),
		gencache.WithPolicy(policy.NewLRU[string, *cacheEntry](policy.WithMaxSize(cfg.MaxCacheSize))),
	)

	r := &Resolver{
		cache:    cache,
		config:   cfg,
		stats:    metrics.NewDNSMetrics(cfg.MaxStatsSize),
		resolver: &net.Resolver{PreferGo: true},
	}

	r.configureResolverDialer()

	return r, nil
}

// LookupIP resolves a hostname to IP addresses with robust retry logic
func (r *Resolver) LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	if host == "" {
		return nil, &net.DNSError{Err: "empty host", Name: host, IsNotFound: true}
	}
	start := time.Now()
	defer func() { r.stats.RecordLookupTime(time.Since(start), map[string]string{}, false) }()

	if ips, err, ok := r.getIPsFromCache(host); ok {
		return ips, err
	}
	r.stats.IncCacheMisses(map[string]string{})
	return r.lookupWithRetry(ctx, host)
}

// GetStats returns current resolver statistics
func (r *Resolver) GetStats() metrics.DNSSnapshot {
	return r.stats.Snapshot()
}

// ClearCache clears the DNS cache
func (r *Resolver) ClearCache() {
	_ = r.cache.Clear() // ignore error
}

// filterIPsByVersion filters IPs based on configured IP version support
func (r *Resolver) filterIPsByVersion(ips []net.IP) []net.IP {
	if (r.config.EnableIPv4 && r.config.EnableIPv6) || len(ips) == 0 {
		return ips
	}

	var filtered []net.IP
	for _, ip := range ips {
		if ip.To4() != nil && r.config.EnableIPv4 {
			filtered = append(filtered, ip)
		} else if ip.To4() == nil && r.config.EnableIPv6 {
			filtered = append(filtered, ip)
		}
	}
	return filtered
}

// isNotFoundError checks if the error is a "not found" DNS error
func isNotFoundError(err error) bool {
	if dnsErr, ok := err.(*net.DNSError); ok {
		return dnsErr.IsNotFound
	}
	return false
}

// calculateBackoff calculates the backoff duration with jitter
func (r *Resolver) calculateBackoff(attempt int) time.Duration {
	baseBackoff := 100 * time.Millisecond
	maxBackoff := 2 * time.Second
	backoff := baseBackoff * time.Duration(1<<uint(attempt))
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	jitter := time.Duration(float64(backoff) * 0.2)
	return backoff - jitter + time.Duration(float64(jitter)*2*rand.Float64())
}

// handleLookupError determines if an error should trigger a retry and returns appropriate error
func (r *Resolver) handleLookupError(err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	// Check for context errors
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		r.stats.IncLookupErrors(map[string]string{})
		if errors.Is(err, context.DeadlineExceeded) {
			return true, err
		}
		return false, err
	}

	// Check for DNS errors
	if dnsErr, ok := err.(*net.DNSError); ok {
		if dnsErr.IsNotFound {
			return false, dnsErr
		}
		if dnsErr.IsTemporary {
			r.stats.IncLookupErrors(map[string]string{})
			return true, dnsErr
		}
		return false, dnsErr
	}

	// Check for network errors
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			r.stats.IncLookupErrors(map[string]string{})
			return true, err
		}
	}

	return false, err
}

// performLookup performs the actual DNS lookup
func (r *Resolver) performLookup(ctx context.Context, host string) ([]net.IP, error) {
	if r.lookupIPFunc != nil {
		return r.lookupIPFunc(ctx, "ip", host)
	}
	return r.resolver.LookupIP(ctx, "ip", host)
}

// getFromCache retrieves a cache entry for the given host
func (r *Resolver) getFromCache(host string) (*cacheEntry, bool) {
	entry, err := r.cache.Get(host)
	if err != nil {
		r.stats.IncCacheErrors(map[string]string{})
		return nil, false
	}
	if entry.ExpiresAt.Before(time.Now()) {
		if err := r.cache.Delete(host); err != nil {
			r.stats.IncCacheErrors(map[string]string{})
		}
		return nil, false
	}
	return entry, true
}

// setCache stores a cache entry for the given host
func (r *Resolver) setCache(host string, ips []net.IP, isNegative bool) {
	ttl := r.config.TTL
	if isNegative {
		ttl = r.config.NegativeTTL
	}

	entry := &cacheEntry{
		IPs:        ips,
		ExpiresAt:  time.Now().Add(ttl),
		IsNegative: isNegative,
	}

	err := r.cache.Set(host, entry, ttl)
	if err != nil {
		r.stats.IncCacheErrors(map[string]string{})
	}
}

func (r *Resolver) getIPsFromCache(host string) ([]net.IP, error, bool) {
	if entry, ok := r.getFromCache(host); ok {
		r.stats.IncCacheHits(map[string]string{})
		if entry.IsNegative {
			return nil, &net.DNSError{Err: "no addresses found (cached negative)", Name: host, IsNotFound: true}, true
		}
		if len(entry.IPs) == 0 {
			return nil, &net.DNSError{Err: "no addresses found", Name: host, IsNotFound: true}, true
		}
		return entry.IPs, nil, true
	}
	return nil, nil, false
}

func (r *Resolver) lookupWithRetry(ctx context.Context, host string) ([]net.IP, error) {
	var ips []net.IP
	var err error
	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			r.stats.RetryAttempts.Add(1)
			backoff := r.calculateBackoff(attempt)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		timeoutCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
		ips, err = r.performLookup(timeoutCtx, host)
		cancel()
		if err == nil {
			break
		}
		shouldRetry, _ := r.handleLookupError(err)
		if !shouldRetry {
			break
		}
		if attempt == 1 && len(r.config.Nameservers) > 1 {
			r.configureResolverDialer()
		}
	}
	if err != nil {
		if isNotFoundError(err) {
			r.setCache(host, nil, true)
		}
		return nil, err
	}
	if len(ips) == 0 {
		r.setCache(host, nil, true)
		return nil, &net.DNSError{Err: "no addresses found", Name: host, IsNotFound: true}
	}
	filteredIPs := r.filterIPsByVersion(ips)
	r.setCache(host, filteredIPs, false)
	return filteredIPs, nil
}

// configureResolverDialer sets up the resolver's dialer function
func (r *Resolver) configureResolverDialer() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.config.Nameservers) == 0 {
		// Use system resolver if no nameservers are configured
		r.resolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
			dialer := net.Dialer{Timeout: r.config.Timeout}
			return dialer.DialContext(ctx, network, address)
		}
		return
	}

	// Configure custom nameservers
	r.resolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		dialer := net.Dialer{Timeout: r.config.Timeout}
		servers := r.config.Nameservers

		// Rotate servers if enabled
		if r.config.RotateServers && len(servers) > 1 {
			idx := int(r.serverIndex.Add(1)) % len(servers)
			servers = append(servers[idx:], servers[:idx]...)
		}

		var firstErr error
		for _, ns := range servers {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ns, strconv.FormatUint(uint64(r.config.Port), 10)))
			if err == nil {
				return conn, nil
			}
			if firstErr == nil {
				firstErr = err
			}
		}
		return nil, fmt.Errorf("failed to connect to any nameserver: %w", firstErr)
	}
}
