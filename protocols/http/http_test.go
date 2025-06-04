package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct {
	resp *http.Response
	err  error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.resp, m.err
}

func newTestTransport(t *testing.T) *HTTPTransport {
	cfg := NewConfig()
	tr, err := NewHTTPTransport(cfg)
	require.NoError(t, err)
	return tr.(*HTTPTransport)
}

func TestNewHTTPTransportConfigError(t *testing.T) {
	cfg := NewConfig()
	cfg.ServerAddress = ""
	_, err := NewHTTPTransport(cfg)
	require.Error(t, err)
}

func TestNewHTTPTransportSuccess(t *testing.T) {
	cfg := NewConfig()
	tr, err := NewHTTPTransport(cfg)
	require.NoError(t, err)
	require.NotNil(t, tr)
}

func TestGetProtocolNameAndVersion(t *testing.T) {
	tr := newTestTransport(t)
	require.Equal(t, "http", tr.GetProtocolName())
	require.Equal(t, "1.1", tr.GetProtocolVersion())
}

func TestSetAndGetTimeout(t *testing.T) {
	tr := newTestTransport(t)
	tr.SetTimeout(42 * time.Second)
	require.Equal(t, 42*time.Second, tr.GetTimeout())
}

func TestSetAndGetMaxRetries(t *testing.T) {
	tr := newTestTransport(t)
	tr.SetMaxRetries(7)
	require.Equal(t, 7, tr.GetMaxRetries())
}

func TestIsConnectedAndConnectionState(t *testing.T) {
	tr := newTestTransport(t)
	tr.state.Connected = true
	require.True(t, tr.IsConnected())
	state := tr.GetConnectionState()
	require.True(t, state.Connected)
}

func TestSetDNSResolver(t *testing.T) {
	tr := newTestTransport(t)
	tr.SetDNSResolver(nil)
	require.Nil(t, tr.resolver)
}

func TestResolveHostError(t *testing.T) {
	tr := newTestTransport(t)
	tr.resolver = nil
	_, err := tr.ResolveHost(context.Background(), "host")
	require.Error(t, err)
}

func TestConnectDisconnect(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = false
	tr.resolver = nil // force error
	err := tr.Connect(context.Background())
	require.Error(t, err)
	err = tr.Disconnect(context.Background())
	require.NoError(t, err)
	tr.closed = true
	err = tr.Disconnect(context.Background())
	require.NoError(t, err)
}

func TestSendReceiveClosed(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = true
	err := tr.Send(context.Background(), []byte("data"))
	require.Error(t, err)
	_, err = tr.Receive(context.Background())
	require.Error(t, err)
}

func TestSendReceiveNotConnected(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = false
	tr.resolver = nil // force error on connect
	err := tr.Send(context.Background(), []byte("data"))
	require.Error(t, err)
	_, err = tr.Receive(context.Background())
	require.Error(t, err)
}

func TestBatchSendBatchReceiveClosed(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = true
	err := tr.BatchSend(context.Background(), [][]byte{{1}, {2}})
	require.Error(t, err)
	_, err = tr.BatchReceive(context.Background(), 2)
	require.Error(t, err)
}

func TestBatchSendBatchReceiveNotConnected(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = false
	tr.resolver = nil // force error on connect
	err := tr.BatchSend(context.Background(), [][]byte{{1}, {2}})
	require.Error(t, err)
	_, err = tr.BatchReceive(context.Background(), 2)
	require.Error(t, err)
}

func TestBatchSendTooLarge(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	big := make([][]byte, 1001)
	err := tr.BatchSend(context.Background(), big)
	require.Error(t, err)
}

func TestBatchReceiveInvalidBatchSize(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	_, err := tr.BatchReceive(context.Background(), 0)
	require.Error(t, err)
	_, err = tr.BatchReceive(context.Background(), 1001)
	require.Error(t, err)
}

func TestCheckResponse(t *testing.T) {
	tr := newTestTransport(t)
	resp := &http.Response{StatusCode: 200}
	require.NoError(t, tr.checkResponse(resp))
	resp.StatusCode = 404
	require.Error(t, tr.checkResponse(resp))
}

func TestSplitBatchResponse(t *testing.T) {
	// Valid batch
	item := []byte("abc")
	var buf bytes.Buffer
	buf.Write([]byte{0, 0, 0, 3})
	buf.Write(item)
	items, err := splitBatchResponse(buf.Bytes(), buf.Len(), 1)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, item, items[0])

	// Invalid item size
	_, err = splitBatchResponse([]byte{0, 0, 0, 255}, 4, 1)
	require.Error(t, err)
}

func TestWithRetry(t *testing.T) {
	tr := newTestTransport(t)
	ctx := context.Background()
	count := 0
	err := tr.withRetry(ctx, "op", func() error {
		count++
		if count < 2 {
			return errors.New("fail")
		}
		return nil
	})
	require.NoError(t, err)
}

func TestWithRetryTimeout(t *testing.T) {
	tr := newTestTransport(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := tr.withRetry(ctx, "op", func() error {
		time.Sleep(20 * time.Millisecond)
		return errors.New("fail")
	})
	require.Error(t, err)
}

func TestUpdateState(t *testing.T) {
	tr := newTestTransport(t)
	tr.updateState(true, nil)
	require.True(t, tr.state.Connected)
	tr.updateState(false, errors.New("fail"))
	require.False(t, tr.state.Connected)
	require.Error(t, tr.state.Error)
}

func TestCheckClosed(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	require.NoError(t, tr.checkClosed())
	tr.closed = true
	require.Error(t, tr.checkClosed())
}

func TestSendRequestCreationError(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	tr.config = NewConfig()
	// Simulate error by setting config to nil
	tr.config = nil
	err := tr.Send(context.Background(), []byte("data"))
	require.Error(t, err)
}

func TestSendExecuteRequestError(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	// Patch client to return error
	tr.client = &http.Client{
		Transport: &mockRoundTripper{resp: nil, err: errors.New("fail")},
	}
	err := tr.Send(context.Background(), []byte("data"))
	require.Error(t, err)
}

func TestSendCheckResponseError(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	tr.client = &http.Client{
		Transport: &mockRoundTripper{resp: &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewReader([]byte("")))}, err: nil},
	}
	err := tr.Send(context.Background(), []byte("data"))
	require.Error(t, err)
}

func TestReceiveReadResponseBodyError(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	tr.client = &http.Client{
		Transport: &mockRoundTripper{resp: &http.Response{StatusCode: 200, Body: errorReadCloser{}}, err: nil},
	}
	_, err := tr.Receive(context.Background())
	require.Error(t, err)
}

type errorReadCloser struct{}

func (errorReadCloser) Read(p []byte) (int, error) { return 0, errors.New("fail read") }
func (errorReadCloser) Close() error               { return nil }

func TestBatchReceiveEmptyResponse(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	tr.client = &http.Client{
		Transport: &mockRoundTripper{resp: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte{}))}, err: nil},
	}
	_, err := tr.BatchReceive(context.Background(), 1)
	require.Error(t, err)
}

func TestBatchReceiveInvalidBatchFormat(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	tr.client = &http.Client{
		Transport: &mockRoundTripper{resp: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte{0, 0, 0, 255}))}, err: nil},
	}
	_, err := tr.BatchReceive(context.Background(), 1)
	require.Error(t, err)
}

func TestWithRetryAllFail(t *testing.T) {
	tr := newTestTransport(t)
	tr.config.MaxRetries = 2
	count := 0
	err := tr.withRetry(context.Background(), "failop", func() error {
		count++
		return errors.New("fail")
	})
	require.Error(t, err)
	require.Equal(t, 2, count)
}

func TestWithRetryContextCancelled(t *testing.T) {
	tr := newTestTransport(t)
	tr.config.MaxRetries = 3
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := tr.withRetry(ctx, "failop", func() error {
		return errors.New("fail")
	})
	require.Error(t, err)
}

func TestConnectAlreadyConnected(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	err := tr.Connect(context.Background())
	require.NoError(t, err)
}

func TestDisconnectAlreadyClosed(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = true
	err := tr.Disconnect(context.Background())
	require.NoError(t, err)
}

func TestBatchSendEdgeCases(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	tr.client = &http.Client{
		Transport: &mockRoundTripper{resp: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte{}))}, err: nil},
	}

	t.Run("empty slice", func(t *testing.T) {
		err := tr.BatchSend(context.Background(), [][]byte{})
		require.NoError(t, err)
	})

	t.Run("nil slice", func(t *testing.T) {
		err := tr.BatchSend(context.Background(), nil)
		require.NoError(t, err)
	})

	t.Run("max size", func(t *testing.T) {
		items := make([][]byte, 1000)
		for i := range items {
			items[i] = []byte("x")
		}
		err := tr.BatchSend(context.Background(), items)
		require.NoError(t, err)
	})
}

func TestBatchReceiveBatchSizes(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = true
	tr.client = &http.Client{
		Transport: &mockRoundTripper{resp: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte{0, 0, 0, 1, 'a'}))}, err: nil},
	}
	_, err := tr.BatchReceive(context.Background(), 1)
	require.NoError(t, err)
}

func TestSplitBatchResponseTableDriven(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		totalSize int
		batchSize int
		wantErr   bool
		wantLen   int
	}{
		{"valid single", []byte{0, 0, 0, 1, 'a'}, 5, 1, false, 1},
		{"valid two", []byte{0, 0, 0, 1, 'a', 0, 0, 0, 1, 'b'}, 10, 2, false, 2},
		{"zero length", []byte{}, 0, 0, false, 0},
		{"partial header", []byte{0, 0, 0}, 3, 1, false, 0},
		{"invalid item size", []byte{0, 0, 0, 255}, 4, 1, true, 0},
		{"item size too large", []byte{0, 0, 0, 10, 'a'}, 5, 1, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := splitBatchResponse(tt.data, tt.totalSize, tt.batchSize)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, items, tt.wantLen)
			}
		})
	}
}

// Simulate pool exhaustion by patching pool.Get to return error
func TestConnectPoolExhausted(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.state.Connected = false
	tr.pool = nil
	tr.client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return nil, errors.New("pool exhausted")
			},
		},
	}
	// Patch config to valid
	tr.config = NewConfig()
	tr.config.ServerAddress = "localhost"
	tr.config.Port = 8080
	tr.resolver = nil // force error in ResolveHost
	err := tr.Connect(context.Background())
	require.Error(t, err)
}

func TestDisconnectNilPool(t *testing.T) {
	tr := newTestTransport(t)
	tr.closed = false
	tr.pool = nil
	tr.state.Connected = true
	err := tr.Disconnect(context.Background())
	require.NoError(t, err)
}
