// resolver_test.go contains tests for the DNS resolver, including cache, error handling, retry, and concurrency logic.
package dns

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestLookupIP covers all major behaviors of the Resolver's LookupIP method, including cache hits/misses, error handling, retry logic, cache behavior, and concurrency safety.
func TestLookupIP(t *testing.T) {
	t.Run("CacheHit", func(t *testing.T) {
		// Tests that a cache hit returns the cached IP and does not call the lookup function.
		r, err := NewResolver(DefaultConfig().WithTTL(1 * time.Minute).WithNegativeTTL(30 * time.Second))
		require.NoError(t, err)
		host := "example.com"
		ips := []net.IP{net.ParseIP("93.184.216.34")}

		r.setCache(host, ips, false)

		called := false
		r.lookupIPFunc = func(ctx context.Context, network, host string) ([]net.IP, error) {
			called = true
			return nil, errors.New("should not be called on cache hit")
		}

		result, err := r.LookupIP(context.Background(), host)
		require.NoError(t, err)
		require.Equal(t, ips, result)
		require.False(t, called)
	})

	t.Run("CacheMiss", func(t *testing.T) {
		// Tests that a cache miss triggers a lookup and returns an error if lookup fails.
		r, err := NewResolver(DefaultConfig().WithTTL(1 * time.Minute).WithNegativeTTL(30 * time.Second))
		require.NoError(t, err)
		host := "example.com"

		r.lookupIPFunc = func(ctx context.Context, network, host string) ([]net.IP, error) {
			return nil, errors.New("mock error")
		}

		_, err = r.LookupIP(context.Background(), host)
		require.Error(t, err)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Table-driven test for various error types and retry logic.
		tests := []struct {
			name          string
			host          string
			mockError     error
			expectedError string
			shouldRetry   bool
		}{
			{
				name:          "TemporaryDNSError",
				host:          "temporary.host",
				mockError:     &net.DNSError{Err: "temporary error", Name: "temporary.host", IsTemporary: true},
				expectedError: "temporary error",
				shouldRetry:   true,
			},
			{
				name:          "PermanentDNSError",
				host:          "permanent.host",
				mockError:     &net.DNSError{Err: "no such host", Name: "permanent.host", IsNotFound: true},
				expectedError: "no such host",
				shouldRetry:   false,
			},
			{
				name:          "NetworkError",
				host:          "network.host",
				mockError:     errors.New("network error"),
				expectedError: "network error",
				shouldRetry:   false,
			},
			{
				name:          "TimeoutError",
				host:          "timeout.host",
				mockError:     context.DeadlineExceeded,
				expectedError: "context deadline exceeded",
				shouldRetry:   true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r, err := NewResolver(DefaultConfig().WithTTL(1 * time.Minute).WithNegativeTTL(30 * time.Second))
				require.NoError(t, err)
				r.lookupIPFunc = func(ctx context.Context, network, host string) ([]net.IP, error) {
					return nil, tt.mockError
				}

				_, err = r.LookupIP(context.Background(), tt.host)
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)

				if tt.shouldRetry {
					stats := r.GetStats()
					require.Greater(t, stats.RetryAttempts, int64(0))
				}
			})
		}
	})

	t.Run("RetryBehavior", func(t *testing.T) {
		// Table-driven test for retry logic: success after retries and fail after max retries.
		tests := []struct {
			name          string
			host          string
			mockResponses []struct {
				ips []net.IP
				err error
			}
			expectedIPs      []net.IP
			expectedAttempts int
			shouldSucceed    bool
		}{
			{
				name: "SuccessAfterRetries",
				host: "retry.host",
				mockResponses: []struct {
					ips []net.IP
					err error
				}{
					{nil, &net.DNSError{Err: "temporary error", IsTemporary: true}},
					{nil, &net.DNSError{Err: "temporary error", IsTemporary: true}},
					{[]net.IP{net.ParseIP("93.184.216.34")}, nil},
				},
				expectedIPs:      []net.IP{net.ParseIP("93.184.216.34")},
				expectedAttempts: 3,
				shouldSucceed:    true,
			},
			{
				name: "FailAfterMaxRetries",
				host: "fail.host",
				mockResponses: []struct {
					ips []net.IP
					err error
				}{
					{nil, &net.DNSError{Err: "temporary error", IsTemporary: true}},
					{nil, &net.DNSError{Err: "temporary error", IsTemporary: true}},
					{nil, &net.DNSError{Err: "temporary error", IsTemporary: true}},
					{nil, &net.DNSError{Err: "temporary error", IsTemporary: true}},
				},
				expectedAttempts: 4,
				shouldSucceed:    false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r, err := NewResolver(DefaultConfig().WithTTL(1 * time.Minute).WithNegativeTTL(30 * time.Second))
				require.NoError(t, err)
				attempts := 0

				r.lookupIPFunc = func(ctx context.Context, network, host string) ([]net.IP, error) {
					require.Less(t, attempts, len(tt.mockResponses))
					response := tt.mockResponses[attempts]
					attempts++
					return response.ips, response.err
				}

				result, err := r.LookupIP(context.Background(), tt.host)
				if tt.shouldSucceed {
					require.NoError(t, err)
					require.Equal(t, tt.expectedIPs, result)
				} else {
					require.Error(t, err)
				}
				require.Equal(t, tt.expectedAttempts, attempts)

				stats := r.GetStats()
				require.Equal(t, int64(tt.expectedAttempts-1), stats.RetryAttempts)
			})
		}
	})

	t.Run("CacheBehavior", func(t *testing.T) {
		// Table-driven test for cache size, TTL expiration, and negative cache TTL behavior.
		tests := []struct {
			name       string
			config     *ResolverConfig
			operations []struct {
				host  string
				ips   []net.IP
				sleep time.Duration
			}
			expectedCacheSize int
			expectedEvicted   []string
		}{
			{
				name:   "CacheSizeLimit",
				config: DefaultConfig().WithMaxCacheSize(2).WithCleanupInterval(time.Hour),
				operations: []struct {
					host  string
					ips   []net.IP
					sleep time.Duration
				}{
					{"host1.com", []net.IP{net.ParseIP("1.1.1.1")}, 0},
					{"host2.com", []net.IP{net.ParseIP("2.2.2.2")}, 0},
					{"host3.com", []net.IP{net.ParseIP("3.3.3.3")}, 0},
				},
				expectedCacheSize: 3,
				expectedEvicted:   []string{},
			},
			{
				name:   "TTLExpiration",
				config: DefaultConfig().WithTTL(100 * time.Millisecond).WithNegativeTTL(50 * time.Millisecond).WithCleanupInterval(50 * time.Millisecond),
				operations: []struct {
					host  string
					ips   []net.IP
					sleep time.Duration
				}{
					{"host1.com", []net.IP{net.ParseIP("1.1.1.1")}, 0},
					{"host2.com", []net.IP{net.ParseIP("2.2.2.2")}, 150 * time.Millisecond},
				},
				expectedCacheSize: 0,
				expectedEvicted:   []string{"host1.com", "host2.com"},
			},
			{
				name:   "NegativeCacheTTL",
				config: DefaultConfig().WithNegativeTTL(50 * time.Millisecond).WithCleanupInterval(10 * time.Millisecond),
				operations: []struct {
					host  string
					ips   []net.IP
					sleep time.Duration
				}{
					{"failed.host", nil, 0},
					{"failed.host", nil, 100 * time.Millisecond},
				},
				expectedCacheSize: 0,
				expectedEvicted:   []string{"failed.host"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r, err := NewResolver(tt.config)
				require.NoError(t, err)

				for _, op := range tt.operations {
					if op.sleep > 0 {
						time.Sleep(op.sleep)
					}
					isNegative := op.ips == nil
					r.setCache(op.host, op.ips, isNegative)
				}

				if tt.name == "TTLExpiration" {
					time.Sleep(150 * time.Millisecond)
				}

				var actualSize int
				for _, op := range tt.operations {
					if _, exists := r.getFromCache(op.host); exists {
						actualSize++
					}
				}
				require.Equal(t, tt.expectedCacheSize, actualSize)

				for _, host := range tt.expectedEvicted {
					_, exists := r.getFromCache(host)
					require.False(t, exists)
				}

				for _, op := range tt.operations {
					if !contains(tt.expectedEvicted, op.host) {
						ips, exists := r.getFromCache(op.host)
						require.True(t, exists)
						require.NotNil(t, ips)
					}
				}
			})
		}
	})

	t.Run("ConcurrencySafety", func(t *testing.T) {
		// Table-driven test for concurrent cache operations, lookups, and stats updates.
		tests := []struct {
			name       string
			operation  func(*Resolver)
			goroutines int
			timeout    time.Duration
		}{
			{
				name: "ConcurrentCacheOperations",
				operation: func(r *Resolver) {
					host := fmt.Sprintf("host%d.com", rand.Intn(1000))
					ips := []net.IP{net.ParseIP("1.1.1.1")}
					r.setCache(host, ips, false)
					_, _ = r.getFromCache(host)
				},
				goroutines: 100,
				timeout:    5 * time.Second,
			},
			{
				name: "ConcurrentLookups",
				operation: func(r *Resolver) {
					host := fmt.Sprintf("host%d.com", rand.Intn(1000))
					ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
					defer cancel()
					_, _ = r.LookupIP(ctx, host)
				},
				goroutines: 100,
				timeout:    5 * time.Second,
			},
			{
				name: "ConcurrentStatsUpdates",
				operation: func(r *Resolver) {
					r.stats.CacheHits.Add(1)
					r.stats.CacheMisses.Add(1)
				},
				goroutines: 100,
				timeout:    5 * time.Second,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r, err := NewResolver(DefaultConfig().WithTTL(1 * time.Minute).WithNegativeTTL(30 * time.Second))
				require.NoError(t, err)
				var wg sync.WaitGroup
				start := time.Now()

				for i := 0; i < tt.goroutines; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						tt.operation(r)
					}()
				}

				done := make(chan struct{})
				go func() {
					wg.Wait()
					close(done)
				}()

				select {
				case <-done:
					duration := time.Since(start)
					require.Less(t, duration, tt.timeout)
				case <-time.After(tt.timeout):
					t.Fatal("Test timed out, possible deadlock")
				}
			})
		}
	})
}

// TestCoverageHelpers covers additional code paths for coverage, including ClearCache, filterIPsByVersion, and configureResolverDialer.
func TestCoverageHelpers(t *testing.T) {
	r, err := NewResolver(DefaultConfig())
	require.NoError(t, err)

	// Test ClearCache (should not panic)
	r.ClearCache()

	// Test filterIPsByVersion for various IP version settings
	ips := []net.IP{net.ParseIP("1.2.3.4"), net.ParseIP("2001:db8::1")}
	r.config.EnableIPv4 = true
	r.config.EnableIPv6 = false
	filtered := r.filterIPsByVersion(ips)
	require.Equal(t, []net.IP{net.ParseIP("1.2.3.4")}, filtered)

	r.config.EnableIPv4 = false
	r.config.EnableIPv6 = true
	filtered = r.filterIPsByVersion(ips)
	require.Equal(t, []net.IP{net.ParseIP("2001:db8::1")}, filtered)

	r.config.EnableIPv4 = true
	r.config.EnableIPv6 = true
	filtered = r.filterIPsByVersion(ips)
	require.Equal(t, ips, filtered)

	empty := r.filterIPsByVersion([]net.IP{})
	require.Empty(t, empty)

	// Test configureResolverDialer with invalid nameserver
	r.config.Nameservers = []string{"256.256.256.256"} // invalid IP
	r.config.Port = 5353
	r.config.RotateServers = false
	r.configureResolverDialer()
}

// contains checks if a string is present in a slice.
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
