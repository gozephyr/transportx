package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorVariables(t *testing.T) {
	testCases := []struct {
		err      error
		expected string
	}{
		{ErrInvalidConfig, "invalid configuration"},
		{ErrNotConnected, "transport not connected"},
		{ErrConnectionFailed, "connection failed"},
		{ErrSendFailed, "send operation failed"},
		{ErrReceiveFailed, "receive operation failed"},
		{ErrBatchSizeExceeded, "batch size exceeds maximum allowed"},
		{ErrInvalidBatchSize, "invalid batch size"},
		{ErrTimeout, "operation timed out"},
		{ErrMaxRetriesExceeded, "maximum retries exceeded"},
		{ErrServerError, "server returned error"},
		{ErrInvalidResponse, "invalid server response"},
		{ErrConnectionClosed, "connection closed"},
		{ErrBufferFull, "buffer is full"},
		{ErrInvalidAddress, "invalid server address"},
		{ErrDNSResolution, "DNS resolution failed"},
		{ErrPoolExhausted, "connection pool exhausted"},
		{ErrInvalidState, "invalid transport state"},
	}

	for _, tc := range testCases {
		require.NotNil(t, tc.err)
		require.EqualError(t, tc.err, tc.expected)
	}
}
