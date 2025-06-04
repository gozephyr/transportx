// Package http provides the HTTP transport implementation for transportx.
package http

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gozephyr/transportx/buffer"
	"github.com/gozephyr/transportx/dns"
	"github.com/gozephyr/transportx/metrics"
	"github.com/gozephyr/transportx/pool"
	"github.com/gozephyr/transportx/protocols"
)

// Constants for batch operations
const (
	batchSizeHeader = "X-Batch-Size"
	maxRedirects    = 10
	maxBatchSize    = 1000
	errFormat       = "%w: %v"
)

// Error messages
const (
	errCreateRequest      = "failed to create request: %w"
	errConnect            = "failed to connect to HTTP server: %w"
	errSendRequest        = "failed to send request: %w"
	errReceiveData        = "failed to receive data: %w"
	errBatchSize          = "batch size exceeds buffer capacity: %d > %d"
	errInvalidBatchSize   = "invalid batch size: %d"
	errReadResponse       = "failed to read response body: %w"
	errReadBatch          = "failed to read batch response body: %w"
	errMissingBatchSize   = "missing batch size in response"
	errInvalidItemSize    = "invalid item size in batch response"
	errEmptyResponse      = "empty response"
	errInvalidBatchFormat = "invalid batch format: %s"
)

// HTTPTransport implements the Protocol interface for HTTP
type HTTPTransport struct {
	config   *Config
	client   *http.Client
	buf      *buffer.Buffer
	resolver *dns.Resolver
	pool     *pool.Pool
	state    protocols.ConnectionState
	metrics  *metrics.Metrics
	mu       sync.RWMutex
	closed   bool
}

// NewHTTPTransport creates a new HTTP transport with optimized settings
func NewHTTPTransport(cfg *Config) (protocols.Protocol, error) {
	if cfg == nil {
		cfg = NewConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Create metrics with detailed latency tracking
	metrics := metrics.NewMetrics(metrics.NewConfig().
		WithLatency(true).
		WithThroughput(true).
		WithErrors(true).
		WithRetries(true).
		WithConnections(true).
		WithActiveStreams(true).
		WithFailedConnects(true))

	// Create DNS resolver with optimized settings
	resolverTimeout := cfg.Timeout
	if resolverTimeout <= 0 {
		resolverTimeout = 30 * time.Second
	}
	resolver, err := dns.NewResolver(&dns.ResolverConfig{
		Timeout:         resolverTimeout,
		TTL:             15 * time.Minute, // Increased TTL
		NegativeTTL:     30 * time.Second, // Set negative TTL
		MaxCacheSize:    5000,             // Increased cache size
		CleanupInterval: 5 * time.Minute,  // Increased cleanup interval
		MaxRetries:      3,                // Add retries for DNS
		MaxStatsSize:    1000,             // Set max stats size
		EnableIPv4:      true,             // Enable IPv4
		EnableIPv6:      true,             // Enable IPv6
		Port:            53,               // Set DNS port
	})
	if err != nil {
		return nil, err
	}

	// Create connection pool with optimized settings
	connPool := pool.NewPool(&pool.Config{
		MaxIdleConns:     cfg.MaxIdleConns * 2,    // Double the idle connections
		MaxActiveConns:   cfg.MaxConnsPerHost * 2, // Double the active connections
		IdleTimeout:      cfg.IdleTimeout / 2,     // Reduce idle timeout
		MaxLifetime:      15 * time.Minute,        // Reduce max lifetime
		MaxWaitDuration:  cfg.MaxWaitDuration,
		HealthCheckDelay: cfg.HealthCheckDelay,
		DialFunc: func(ctx context.Context) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   cfg.Timeout,
				KeepAlive: cfg.KeepAlive,
				DualStack: true,
			}
			addr := fmt.Sprintf("%s:%d", cfg.ServerAddress, cfg.Port)
			return dialer.DialContext(ctx, "tcp", addr)
		},
	})

	// Create custom transport with optimized settings
	transport := &http.Transport{
		MaxIdleConns:          cfg.MaxIdleConns * 2,
		MaxIdleConnsPerHost:   cfg.MaxConnsPerHost * 2,
		IdleConnTimeout:       cfg.IdleTimeout / 2,
		ForceAttemptHTTP2:     true, // Always attempt HTTP/2
		MaxConnsPerHost:       cfg.MaxConnsPerHost * 2,
		ResponseHeaderTimeout: cfg.ResponseTimeout,
		ExpectContinueTimeout: 500 * time.Millisecond, // Reduced timeout
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := connPool.Get(ctx)
			if err != nil {
				return nil, fmt.Errorf(errFormat, ErrPoolExhausted, err)
			}
			return conn, nil
		},
		DisableCompression:  false,                  // Enable compression
		DisableKeepAlives:   false,                  // Enable keep-alives
		WriteBufferSize:     cfg.MaxHeaderBytes * 2, // Increased buffer size
		ReadBufferSize:      cfg.MaxHeaderBytes * 2, // Increased buffer size
		TLSHandshakeTimeout: 5 * time.Second,        // Reduced TLS handshake timeout
	}

	// Create client with optimized settings
	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}

	return &HTTPTransport{
		config:   cfg,
		client:   client,
		buf:      buffer.NewBuffer(),
		resolver: resolver,
		pool:     connPool,
		state: protocols.ConnectionState{
			Connected:    false,
			LastActivity: time.Now(),
			Metadata:     make(map[string]any),
		},
		metrics: metrics,
	}, nil
}

// GetProtocolName returns the name of the protocol
func (t *HTTPTransport) GetProtocolName() string {
	return "http"
}

// GetProtocolVersion returns the version of the protocol
func (t *HTTPTransport) GetProtocolVersion() string {
	return "1.1"
}

// SetTimeout sets the timeout for operations
func (t *HTTPTransport) SetTimeout(timeout time.Duration) {
	t.config.Timeout = timeout
}

// GetTimeout returns the current timeout
func (t *HTTPTransport) GetTimeout() time.Duration {
	return t.config.Timeout
}

// SetMaxRetries sets the maximum number of retries
func (t *HTTPTransport) SetMaxRetries(retries int) {
	t.config.MaxRetries = retries
}

// GetMaxRetries returns the maximum number of retries
func (t *HTTPTransport) GetMaxRetries() int {
	return t.config.MaxRetries
}

// IsConnected returns whether the transport is connected
func (t *HTTPTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state.Connected
}

// GetConnectionState returns the current connection state
func (t *HTTPTransport) GetConnectionState() protocols.ConnectionState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

// SetDNSResolver sets a custom DNS resolver
func (t *HTTPTransport) SetDNSResolver(resolver *dns.Resolver) {
	t.resolver = resolver
}

// ResolveHost resolves a hostname to IP addresses
func (t *HTTPTransport) ResolveHost(ctx context.Context, host string) ([]string, error) {
	if t.resolver == nil {
		return nil, errors.New("resolver is nil")
	}
	ips, err := t.resolver.LookupIP(ctx, host)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(ips))
	for i, ip := range ips {
		result[i] = ip.String()
	}
	return result, nil
}

// Connect establishes a connection to the HTTP server
func (t *HTTPTransport) Connect(ctx context.Context) error {
	if err := t.checkClosed(); err != nil {
		return err
	}

	t.mu.Lock()
	if t.state.Connected {
		t.mu.Unlock()
		return nil // Already connected
	}
	t.mu.Unlock()

	cfg := t.config
	host := cfg.ServerAddress
	if ips, err := t.ResolveHost(ctx, host); err == nil && len(ips) > 0 {
		host = ips[0]
	}

	err := t.withRetry(ctx, "connect", func() error {
		return t.doConnect(ctx, host)
	})
	if err != nil {
		t.updateState(false, err)
		return fmt.Errorf(errFormat, ErrConnectionFailed, err)
	}

	t.updateState(true, nil)
	return nil
}

// Disconnect closes the HTTP client and connection pool
func (t *HTTPTransport) Disconnect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil // Already closed
	}

	t.closed = true
	t.state.Connected = false
	t.state.LastActivity = time.Now()

	if t.pool != nil {
		t.pool.Close()
	}

	return nil
}

// Send sends data to the HTTP server
func (t *HTTPTransport) Send(ctx context.Context, data []byte) error {
	if err := t.checkClosed(); err != nil {
		return err
	}

	if !t.IsConnected() {
		if err := t.Connect(ctx); err != nil {
			return err
		}
	}

	req, err := t.createRequest(ctx, http.MethodPost, data)
	if err != nil {
		return fmt.Errorf(errFormat, ErrSendFailed, err)
	}

	resp, err := t.executeRequest(ctx, req)
	if err != nil {
		return fmt.Errorf(errFormat, ErrSendFailed, err)
	}
	defer resp.Body.Close()

	if err := t.checkResponse(resp); err != nil {
		return fmt.Errorf(errFormat, ErrSendFailed, err)
	}

	t.updateState(true, nil)
	return nil
}

// Receive receives data from the HTTP server
func (t *HTTPTransport) Receive(ctx context.Context) ([]byte, error) {
	if err := t.checkClosed(); err != nil {
		return nil, err
	}

	if !t.IsConnected() {
		if err := t.Connect(ctx); err != nil {
			return nil, err
		}
	}

	req, err := t.createRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf(errFormat, ErrReceiveFailed, err)
	}

	resp, err := t.executeRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf(errFormat, ErrReceiveFailed, err)
	}
	defer resp.Body.Close()

	if err := t.checkResponse(resp); err != nil {
		return nil, fmt.Errorf(errFormat, ErrReceiveFailed, err)
	}

	data, err := t.readResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf(errFormat, ErrReceiveFailed, err)
	}

	t.updateState(true, nil)
	return data, nil
}

// BatchSend sends multiple data items in a single request
func (t *HTTPTransport) BatchSend(ctx context.Context, dataItems [][]byte) error {
	if err := t.checkClosed(); err != nil {
		return err
	}

	if !t.IsConnected() {
		if err := t.Connect(ctx); err != nil {
			return err
		}
	}

	if len(dataItems) > maxBatchSize {
		return fmt.Errorf(errBatchSize, len(dataItems), maxBatchSize)
	}

	// Prepare batch data
	var batchData []byte
	for _, item := range dataItems {
		sizeBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(sizeBytes, uint32(len(item)))
		batchData = append(batchData, sizeBytes...)
		batchData = append(batchData, item...)
	}

	req, err := t.createRequest(ctx, http.MethodPost, batchData)
	if err != nil {
		return fmt.Errorf(errFormat, ErrSendFailed, err)
	}
	req.Header.Set(batchSizeHeader, strconv.Itoa(len(dataItems)))

	resp, err := t.executeRequest(ctx, req)
	if err != nil {
		return fmt.Errorf(errFormat, ErrSendFailed, err)
	}
	defer resp.Body.Close()

	if err := t.checkResponse(resp); err != nil {
		return fmt.Errorf(errFormat, ErrSendFailed, err)
	}

	t.updateState(true, nil)
	return nil
}

// BatchReceive receives multiple data items in a single request
func (t *HTTPTransport) BatchReceive(ctx context.Context, batchSize int) ([][]byte, error) {
	if err := t.checkClosed(); err != nil {
		return nil, err
	}

	if !t.IsConnected() {
		if err := t.Connect(ctx); err != nil {
			return nil, err
		}
	}

	if batchSize <= 0 || batchSize > maxBatchSize {
		return nil, fmt.Errorf(errInvalidBatchSize, batchSize)
	}

	req, err := t.createRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf(errFormat, ErrReceiveFailed, err)
	}
	req.Header.Set(batchSizeHeader, strconv.Itoa(batchSize))

	resp, err := t.executeRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf(errFormat, ErrReceiveFailed, err)
	}
	defer resp.Body.Close()

	if err := t.checkResponse(resp); err != nil {
		return nil, fmt.Errorf(errFormat, ErrReceiveFailed, err)
	}

	data, err := t.readResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf(errFormat, ErrReceiveFailed, err)
	}

	if len(data) == 0 {
		return nil, errors.New(errEmptyResponse)
	}

	items, err := splitBatchResponse(data, len(data), batchSize)
	if err != nil {
		return nil, fmt.Errorf(errFormat, ErrReceiveFailed, err)
	}

	t.updateState(true, nil)
	return items, nil
}

// GetMetrics returns the transport metrics
func (t *HTTPTransport) GetMetrics() *metrics.Metrics {
	return t.metrics
}

// createRequest creates a new HTTP request
func (t *HTTPTransport) createRequest(ctx context.Context, method string, data []byte) (*http.Request, error) {
	if t.config == nil {
		return nil, errors.New("config is nil")
	}
	url := t.config.GetFullURL()
	var body io.Reader
	if data != nil {
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf(errCreateRequest, err)
	}
	return req, nil
}

// executeRequest executes an HTTP request
func (t *HTTPTransport) executeRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(errSendRequest, err)
	}
	return resp, nil
}

// readResponseBody reads the response body
func (t *HTTPTransport) readResponseBody(resp *http.Response) ([]byte, error) {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(errReadResponse, err)
	}
	return data, nil
}

// withRetry executes an operation with optimized retry logic
func (t *HTTPTransport) withRetry(ctx context.Context, operation string, fn func() error) error {
	var lastErr error
	backoff := time.Second

	for i := 0; i < t.config.MaxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		// Implement exponential backoff
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2 // Exponential backoff
			if backoff > 5*time.Second {
				backoff = 5 * time.Second // Cap maximum backoff
			}
			continue
		}
	}
	return lastErr
}

// checkResponse checks if the response is valid
func (t *HTTPTransport) checkResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

// splitBatchResponse splits a batch response into individual items
func splitBatchResponse(data []byte, totalSize, batchSize int) ([][]byte, error) {
	items := make([][]byte, 0, batchSize)
	offset := 0

	for len(items) < batchSize && offset+4 <= totalSize {
		itemLen := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4

		if itemLen <= 0 || offset+itemLen > totalSize {
			return nil, errors.New(errInvalidItemSize)
		}

		items = append(items, data[offset:offset+itemLen])
		offset += itemLen
	}

	return items, nil
}

// updateState updates the connection state
func (t *HTTPTransport) updateState(connected bool, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.state.Connected = connected
	t.state.LastActivity = time.Now()
	t.state.Error = err
}

// doConnect performs the actual connection
func (t *HTTPTransport) doConnect(ctx context.Context, host string) error {
	req, err := t.createRequest(ctx, http.MethodGet, nil)
	if err != nil {
		return err
	}

	resp, err := t.executeRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return t.checkResponse(resp)
}

// checkClosed checks if the transport is closed
func (t *HTTPTransport) checkClosed() error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.closed {
		return errors.New("transport is closed")
	}
	return nil
}
