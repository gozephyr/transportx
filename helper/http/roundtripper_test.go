package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockHTTPTransport struct {
	sendCalled    bool
	receiveCalled bool
	sendData      []byte
	respData      []byte
	sendErr       error
	receiveErr    error
}

func (m *mockHTTPTransport) Send(ctx context.Context, data []byte) error {
	m.sendCalled = true
	m.sendData = data
	return m.sendErr
}

func (m *mockHTTPTransport) Receive(ctx context.Context) ([]byte, error) {
	m.receiveCalled = true
	return m.respData, m.receiveErr
}

func TestTransportxRoundTripperRoundtripSuccess(t *testing.T) {
	mock := &mockHTTPTransport{respData: []byte("response")}
	rt := &TransportxRoundTripper{transport: mock}

	req, err := http.NewRequest("POST", "http://example.com", bytes.NewReader([]byte("input")))
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.True(t, mock.sendCalled, "Send should be called")
	require.True(t, mock.receiveCalled, "Receive should be called")
	require.Equal(t, []byte("input"), mock.sendData)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, []byte("response"), body)
}

func TestTransportxRoundTripperRoundtripSendError(t *testing.T) {
	mock := &mockHTTPTransport{sendErr: context.DeadlineExceeded}
	rt := &TransportxRoundTripper{transport: mock}

	req, err := http.NewRequest("POST", "http://example.com", bytes.NewReader([]byte("input")))
	require.NoError(t, err)
	_, err = rt.RoundTrip(req)
	require.Error(t, err)
}

func TestTransportxRoundTripperRoundtripReceiveError(t *testing.T) {
	mock := &mockHTTPTransport{receiveErr: context.DeadlineExceeded}
	rt := &TransportxRoundTripper{transport: mock}

	req, err := http.NewRequest("POST", "http://example.com", bytes.NewReader([]byte("input")))
	require.NoError(t, err)
	_, err = rt.RoundTrip(req)
	require.Error(t, err)
}
