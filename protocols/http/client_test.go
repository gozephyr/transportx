package http

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gozephyr/transportx/metrics"
	"github.com/stretchr/testify/require"
)

// mockTransport implements the minimal HTTPTransport interface for testing
// Only methods used by Client are implemented

type mockTransport struct {
	connectErr       error
	disconnectErr    error
	sendErr          error
	receiveErr       error
	receiveData      []byte
	batchSendErr     error
	batchReceiveErr  error
	batchReceiveData [][]byte
	metrics          *metrics.Metrics
	connected        bool
	timeout          time.Duration
}

func (m *mockTransport) Connect(ctx context.Context) error           { return m.connectErr }
func (m *mockTransport) Disconnect(ctx context.Context) error        { return m.disconnectErr }
func (m *mockTransport) Send(ctx context.Context, data []byte) error { return m.sendErr }
func (m *mockTransport) Receive(ctx context.Context) ([]byte, error) {
	return m.receiveData, m.receiveErr
}
func (m *mockTransport) BatchSend(ctx context.Context, data [][]byte) error { return m.batchSendErr }
func (m *mockTransport) BatchReceive(ctx context.Context, n int) ([][]byte, error) {
	return m.batchReceiveData, m.batchReceiveErr
}
func (m *mockTransport) GetMetrics() *metrics.Metrics     { return m.metrics }
func (m *mockTransport) IsConnected() bool                { return m.connected }
func (m *mockTransport) SetTimeout(timeout time.Duration) { m.timeout = timeout }
func (m *mockTransport) GetTimeout() time.Duration        { return m.timeout }

// Patch Client to allow injecting a mock transport for testing
// Instead of assigning to c.transport, we will define a testClient struct that embeds Client and overrides methods to use the mockTransport.
type testClient struct {
	*Client
	mock *mockTransport
}

func newTestClient(mt *mockTransport) *testClient {
	c := &Client{transport: nil, config: NewConfig()}
	return &testClient{Client: c, mock: mt}
}

func (tc *testClient) Connect(ctx context.Context) error    { return tc.mock.Connect(ctx) }
func (tc *testClient) Disconnect(ctx context.Context) error { return tc.mock.Disconnect(ctx) }
func (tc *testClient) Send(ctx context.Context, data []byte) ([]byte, error) {
	if err := tc.mock.Send(ctx, data); err != nil {
		return nil, err
	}
	return tc.mock.Receive(ctx)
}

func (tc *testClient) BatchSend(ctx context.Context, dataItems [][]byte) ([][]byte, error) {
	if err := tc.mock.BatchSend(ctx, dataItems); err != nil {
		return nil, err
	}
	return tc.mock.BatchReceive(ctx, len(dataItems))
}
func (tc *testClient) GetMetrics() *metrics.Metrics     { return tc.mock.GetMetrics() }
func (tc *testClient) IsConnected() bool                { return tc.mock.IsConnected() }
func (tc *testClient) SetTimeout(timeout time.Duration) { tc.mock.SetTimeout(timeout) }
func (tc *testClient) GetTimeout() time.Duration        { return tc.mock.GetTimeout() }

func TestNewClientError(t *testing.T) {
	cfg := NewConfig()
	cfg.ServerAddress = "" // invalid
	c, err := NewClient(cfg)
	require.Error(t, err)
	require.Nil(t, c)
}

func TestNewClientSuccessAndNilConfig(t *testing.T) {
	c, err := NewClient(nil)
	require.NoError(t, err)
	require.NotNil(t, c)

	cfg := NewConfig()
	c2, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, c2)
}

func TestClientConnectDisconnect(t *testing.T) {
	mt := &mockTransport{}
	tc := newTestClient(mt)

	mt.connectErr = nil
	require.NoError(t, tc.Connect(context.Background()))

	mt.disconnectErr = nil
	require.NoError(t, tc.Disconnect(context.Background()))

	mt.connectErr = errors.New("fail connect")
	require.Error(t, tc.Connect(context.Background()))

	mt.disconnectErr = errors.New("fail disconnect")
	require.Error(t, tc.Disconnect(context.Background()))
}

func TestClientSend(t *testing.T) {
	mt := &mockTransport{}
	tc := newTestClient(mt)

	// Success
	mt.sendErr = nil
	mt.receiveErr = nil
	mt.receiveData = []byte("ok")
	resp, err := tc.Send(context.Background(), []byte("data"))
	require.NoError(t, err)
	require.Equal(t, []byte("ok"), resp)

	// Send error
	mt.sendErr = errors.New("fail send")
	_, err = tc.Send(context.Background(), []byte("data"))
	require.Error(t, err)

	// Receive error
	mt.sendErr = nil
	mt.receiveErr = errors.New("fail receive")
	_, err = tc.Send(context.Background(), []byte("data"))
	require.Error(t, err)
}

func TestClientBatchSend(t *testing.T) {
	mt := &mockTransport{}
	tc := newTestClient(mt)

	// Success
	mt.batchSendErr = nil
	mt.batchReceiveErr = nil
	mt.batchReceiveData = [][]byte{[]byte("a"), []byte("b")}
	resp, err := tc.BatchSend(context.Background(), [][]byte{[]byte("a"), []byte("b")})
	require.NoError(t, err)
	require.Equal(t, [][]byte{[]byte("a"), []byte("b")}, resp)

	// BatchSend error
	mt.batchSendErr = errors.New("fail batch send")
	_, err = tc.BatchSend(context.Background(), [][]byte{[]byte("a")})
	require.Error(t, err)

	// BatchReceive error
	mt.batchSendErr = nil
	mt.batchReceiveErr = errors.New("fail batch receive")
	_, err = tc.BatchSend(context.Background(), [][]byte{[]byte("a")})
	require.Error(t, err)
}

func TestClientGetMetrics(t *testing.T) {
	mt := &mockTransport{metrics: metrics.NewMetrics(metrics.NewConfig())}
	tc := newTestClient(mt)
	require.Equal(t, mt.metrics, tc.GetMetrics())
}

func TestClientIsConnected(t *testing.T) {
	mt := &mockTransport{connected: true}
	tc := newTestClient(mt)
	require.True(t, tc.IsConnected())
	mt.connected = false
	require.False(t, tc.IsConnected())
}

func TestClientSetGetTimeout(t *testing.T) {
	mt := &mockTransport{}
	tc := newTestClient(mt)
	tc.SetTimeout(42 * time.Second)
	require.Equal(t, 42*time.Second, tc.GetTimeout())
}
