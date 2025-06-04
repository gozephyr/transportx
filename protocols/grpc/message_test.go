package grpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageStruct(t *testing.T) {
	msg := &Message{Data: []byte("hello")}
	require.Equal(t, []byte("hello"), msg.Data)
	msg.Data = []byte("world")
	require.Equal(t, []byte("world"), msg.Data)
}
