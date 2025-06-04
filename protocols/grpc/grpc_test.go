package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewGRPCTransportDefaultConfig(t *testing.T) {
	tr, err := NewGRPCTransport(nil)
	require.NoError(t, err)
	require.NotNil(t, tr)
}

func TestNewGRPCTransportInvalidConfig(t *testing.T) {
	cfg := NewConfig()
	cfg.ServerAddress = "" // Invalid
	_, err := NewGRPCTransport(cfg)
	require.Error(t, err)
}

func TestGRPCTransportConnectDisconnectError(t *testing.T) {
	cfg := NewConfig()
	cfg.ServerAddress = "localhost"
	cfg.Port = 65535 // Unlikely to be open
	tr, err := NewGRPCTransport(cfg)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err = tr.Connect(ctx)
	require.Error(t, err)
	err = tr.Disconnect(ctx)
	require.NoError(t, err)
}

func TestGRPCTransportMethodsNotConnected(t *testing.T) {
	cfg := NewConfig()
	cfg.ServerAddress = "localhost"
	cfg.Port = 65535
	tr, err := NewGRPCTransport(cfg)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	// Should error because not connected
	err = tr.Send(ctx, []byte("data"))
	require.Error(t, err)
	_, err = tr.Receive(ctx)
	require.Error(t, err)
	grpcTr := tr.(*GRPCTransport)
	err = grpcTr.BatchSend(ctx, [][]byte{{'a'}})
	require.Error(t, err)
	_, err = grpcTr.BatchReceive(ctx, 1)
	require.Error(t, err)
}

func TestServerConstructor(t *testing.T) {
	s := NewServer("localhost:0")
	require.NotNil(t, s)
}

func TestServerRPCStubs(t *testing.T) {
	s := NewServer(":0")
	ctx := context.Background()
	msg, err := s.Send(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, msg)
	msg, err = s.Receive(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, msg)
	_, err = s.BatchSend(ctx, nil)
	require.NoError(t, err)
	_, err = s.BatchReceive(ctx, nil)
	require.NoError(t, err)
}
