package grpc

import (
	"context"
	"testing"

	generated "github.com/gozephyr/transportx/proto/generated/proto"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	s := NewServer(":0")
	require.NotNil(t, s)
}

func TestServerStopMultiple(t *testing.T) {
	s := NewServer(":0")
	s.Stop()
	s.Stop() // Should not panic
}

func TestServerRPCs(t *testing.T) {
	s := NewServer(":0")
	ctx := context.Background()
	msg, err := s.Send(ctx, &generated.Message{Data: []byte("foo")})
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), msg.GetData())
	msg, err = s.Receive(ctx, &generated.Empty{})
	require.NoError(t, err)
	require.NotNil(t, msg)
	_, err = s.BatchSend(ctx, &generated.BatchMessage{})
	require.NoError(t, err)
	_, err = s.BatchReceive(ctx, &generated.BatchRequest{BatchSize: 2})
	require.NoError(t, err)
	// StreamMessages: not tested here (requires streaming mock)
}
