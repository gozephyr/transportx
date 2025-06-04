// Package pool provides connection pooling functionality.
package pool

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gozephyr/transportx/metrics"
)

// Default pool settings
const (
	defaultMaxIdleConns     = 100
	defaultMaxActiveConns   = 1000
	defaultIdleTimeout      = 90 * time.Second
	defaultMaxLifetime      = 30 * time.Minute
	defaultMaxWaitDuration  = 5 * time.Second
	defaultHealthCheckDelay = 30 * time.Second
)

// Config holds configuration for the connection pool
type Config struct {
	// Connection limits
	MaxIdleConns     int
	MaxActiveConns   int
	IdleTimeout      time.Duration
	MaxLifetime      time.Duration
	MaxWaitDuration  time.Duration
	HealthCheckDelay time.Duration
	DialFunc         func(context.Context) (net.Conn, error)
}

// NewConfig creates a new pool configuration with default values
func NewConfig() *Config {
	return &Config{
		MaxIdleConns:     defaultMaxIdleConns,
		MaxActiveConns:   defaultMaxActiveConns,
		IdleTimeout:      defaultIdleTimeout,
		MaxLifetime:      defaultMaxLifetime,
		MaxWaitDuration:  defaultMaxWaitDuration,
		HealthCheckDelay: defaultHealthCheckDelay,
	}
}

// Validate validates the pool configuration
func (c *Config) Validate() error {
	if c.MaxIdleConns <= 0 {
		return errors.New("max idle connections must be positive")
	}
	if c.MaxActiveConns <= 0 {
		return errors.New("max active connections must be positive")
	}
	if c.IdleTimeout <= 0 {
		return errors.New("idle timeout must be positive")
	}
	if c.MaxLifetime <= 0 {
		return errors.New("max lifetime must be positive")
	}
	if c.MaxWaitDuration <= 0 {
		return errors.New("max wait duration must be positive")
	}
	if c.HealthCheckDelay <= 0 {
		return errors.New("health check delay must be positive")
	}
	if c.DialFunc == nil {
		return errors.New("dial function must be provided")
	}
	return nil
}

// WithMaxIdleConns sets the maximum number of idle connections
func (c *Config) WithMaxIdleConns(n int) *Config {
	c.MaxIdleConns = n
	return c
}

// WithMaxActiveConns sets the maximum number of active connections
func (c *Config) WithMaxActiveConns(n int) *Config {
	c.MaxActiveConns = n
	return c
}

// WithIdleTimeout sets the idle timeout
func (c *Config) WithIdleTimeout(d time.Duration) *Config {
	c.IdleTimeout = d
	return c
}

// WithMaxLifetime sets the maximum connection lifetime
func (c *Config) WithMaxLifetime(d time.Duration) *Config {
	c.MaxLifetime = d
	return c
}

// WithMaxWaitDuration sets the maximum wait duration
func (c *Config) WithMaxWaitDuration(d time.Duration) *Config {
	c.MaxWaitDuration = d
	return c
}

// WithHealthCheckDelay sets the health check delay
func (c *Config) WithHealthCheckDelay(d time.Duration) *Config {
	c.HealthCheckDelay = d
	return c
}

// WithDialFunc sets the dial function
func (c *Config) WithDialFunc(fn func(context.Context) (net.Conn, error)) *Config {
	c.DialFunc = fn
	return c
}

// Conn represents a pooled connection
type Conn struct {
	net.Conn
	pool     *Pool
	created  time.Time
	lastUsed time.Time
}

// Pool manages a pool of connections
type Pool struct {
	config    *Config
	mu        sync.RWMutex
	idleConns chan *Conn
	closed    atomic.Bool
	metrics   *metrics.PoolMetrics
	done      chan struct{}
}

// NewPool creates a new connection pool with the given configuration
func NewPool(config *Config) *Pool {
	if config == nil {
		config = NewConfig()
	}

	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid pool configuration: %v", err))
	}

	pool := &Pool{
		config:    config,
		idleConns: make(chan *Conn, config.MaxIdleConns),
		metrics:   metrics.NewPoolMetrics(),
		done:      make(chan struct{}),
	}
	pool.metrics.SetMaxActiveConns(int64(config.MaxActiveConns), map[string]string{})
	pool.metrics.SetMaxIdleConns(int64(config.MaxIdleConns), map[string]string{})

	go pool.healthCheck()
	return pool
}

// Get acquires a connection from the pool
func (p *Pool) Get(ctx context.Context) (*Conn, error) {
	if p.closed.Load() {
		return nil, errors.New("pool is closed")
	}

	// Try to get an idle connection
	select {
	case conn := <-p.idleConns:
		if conn.isExpired() {
			conn.Close()
			return p.createConn(ctx)
		}
		p.metrics.IncActiveConns(map[string]string{})
		p.metrics.DecIdleConns(map[string]string{})
		return conn, nil
	default:
		// No idle connections available
		if p.metrics.ActiveConns.Load() >= int64(p.config.MaxActiveConns) {
			// Wait for a connection to become available
			p.metrics.IncWaitingConns(map[string]string{})
			defer p.metrics.DecWaitingConns(map[string]string{})

			select {
			case conn := <-p.idleConns:
				if conn.isExpired() {
					conn.Close()
					return p.createConn(ctx)
				}
				p.metrics.IncActiveConns(map[string]string{})
				p.metrics.DecIdleConns(map[string]string{})
				return conn, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(p.config.MaxWaitDuration):
				return nil, errors.New("connection wait timeout")
			}
		}

		return p.createConn(ctx)
	}
}

// Put returns a connection to the pool
func (p *Pool) Put(conn *Conn) {
	if p.closed.Load() {
		conn.Close()
		return
	}

	p.metrics.DecActiveConns(map[string]string{})
	conn.lastUsed = time.Now()

	select {
	case p.idleConns <- conn:
		p.metrics.IncIdleConns(map[string]string{})
		// Connection returned to pool
	default:
		// Pool is full, close the connection
		conn.Close()
	}
}

// Close closes the pool and all its connections
func (p *Pool) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil
	}

	close(p.done)

	// Close all idle connections
	p.mu.Lock()
	close(p.idleConns)
	for conn := range p.idleConns {
		conn.Close()
	}
	p.mu.Unlock()

	return nil
}

// Stats returns the current pool statistics
func (p *Pool) Stats() metrics.PoolSnapshot {
	return p.metrics.Snapshot()
}

// createConn creates a new connection
func (p *Pool) createConn(ctx context.Context) (*Conn, error) {
	conn, err := p.config.DialFunc(ctx)
	if err != nil {
		return nil, err
	}

	p.metrics.IncActiveConns(map[string]string{})
	p.metrics.IncTotalConns(map[string]string{})
	return &Conn{
		Conn:     conn,
		pool:     p,
		created:  time.Now(),
		lastUsed: time.Now(),
	}, nil
}

// healthCheck periodically checks and cleans up expired connections
func (p *Pool) healthCheck() {
	ticker := time.NewTicker(p.config.HealthCheckDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.cleanup()
		case <-p.done:
			return
		}
	}
}

// cleanup removes expired connections from the pool
func (p *Pool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed.Load() {
		return
	}

	// Create new channel for non-expired connections
	newIdleConns := make(chan *Conn, p.config.MaxIdleConns)
	oldIdleConns := p.idleConns
	p.idleConns = newIdleConns

	// Process existing connections
	for {
		select {
		case conn, ok := <-oldIdleConns:
			if !ok {
				return
			}
			if conn.isExpired() {
				conn.Close()
				p.metrics.DecActiveConns(map[string]string{})
				p.metrics.DecIdleConns(map[string]string{})
			} else {
				select {
				case newIdleConns <- conn:
					// keep idle count the same
				default:
					conn.Close()
					p.metrics.DecIdleConns(map[string]string{})
				}
			}
		default:
			return
		}
	}
}

// isExpired checks if a connection has expired
func (c *Conn) isExpired() bool {
	now := time.Now()
	return now.Sub(c.lastUsed) > c.pool.config.IdleTimeout ||
		now.Sub(c.created) > c.pool.config.MaxLifetime
}
