package pool

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// mockConn is a simple mock for net.Conn
// Only implements the methods required for testing

type mockConn struct {
	closed int32
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConn) Close() error                       { atomic.StoreInt32(&m.closed, 1); return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }
func (m *mockConn) IsClosed() bool                     { return atomic.LoadInt32(&m.closed) == 1 }

func newMockDialFunc() func(context.Context) (net.Conn, error) {
	return func(ctx context.Context) (net.Conn, error) {
		return &mockConn{}, nil
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := NewConfig()
	cfg.DialFunc = newMockDialFunc()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}

	cfg.MaxIdleConns = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for MaxIdleConns = 0")
	}
}

func TestPoolGetPut(t *testing.T) {
	cfg := NewConfig().WithDialFunc(newMockDialFunc())
	pool := NewPool(cfg)
	t.Cleanup(func() { pool.Close() })

	ctx := context.Background()
	conn, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}

	pool.Put(conn)
	stats := pool.Stats()
	if stats.IdleConns != 1 {
		t.Errorf("expected 1 idle conn, got %d", stats.IdleConns)
	}
}

func TestPoolMaxActiveConns(t *testing.T) {
	cfg := NewConfig().WithMaxActiveConns(1).WithDialFunc(newMockDialFunc())
	pool := NewPool(cfg)
	t.Cleanup(func() { pool.Close() })

	ctx := context.Background()
	conn1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("failed to get first connection: %v", err)
	}

	// Try to get a second connection, should block and timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	_, err = pool.Get(ctxTimeout)
	if err == nil {
		t.Error("expected error due to max active conns reached")
	}
	pool.Put(conn1)
}

func TestPoolClose(t *testing.T) {
	cfg := NewConfig().WithDialFunc(newMockDialFunc())
	pool := NewPool(cfg)

	ctx := context.Background()
	conn, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}
	pool.Put(conn)

	if err := pool.Close(); err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}

	if _, err := pool.Get(ctx); err == nil {
		t.Error("expected error after pool is closed")
	}
}

func TestConnIsExpired(t *testing.T) {
	cfg := NewConfig().WithIdleTimeout(10 * time.Millisecond).WithMaxLifetime(20 * time.Millisecond).WithDialFunc(newMockDialFunc())
	pool := NewPool(cfg)
	t.Cleanup(func() { pool.Close() })

	ctx := context.Background()
	conn, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}
	conn.lastUsed = time.Now().Add(-15 * time.Millisecond)
	if !conn.isExpired() {
		t.Error("expected connection to be expired by idle timeout")
	}
	conn.lastUsed = time.Now()
	conn.created = time.Now().Add(-25 * time.Millisecond)
	if !conn.isExpired() {
		t.Error("expected connection to be expired by max lifetime")
	}
}

func TestDialFuncError(t *testing.T) {
	cfg := NewConfig().WithDialFunc(func(ctx context.Context) (net.Conn, error) {
		return nil, errors.New("dial error")
	})
	pool := NewPool(cfg)
	t.Cleanup(func() { pool.Close() })

	ctx := context.Background()
	_, err := pool.Get(ctx)
	if err == nil {
		t.Error("expected error from dial func")
	}
}
