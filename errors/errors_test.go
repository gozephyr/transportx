package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorVariables(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		expect string
	}{
		{"ErrUnsupportedTransport", ErrUnsupportedTransport, "unsupported transport type"},
		{"ErrInvalidConfig", ErrInvalidConfig, "invalid configuration"},
		{"ErrNotConnected", ErrNotConnected, "transport not connected"},
		{"ErrConnectionFailed", ErrConnectionFailed, "connection failed"},
		{"ErrTimeout", ErrTimeout, "operation timed out"},
		{"ErrBufferFull", ErrBufferFull, "buffer is full"},
		{"ErrBufferEmpty", ErrBufferEmpty, "buffer is empty"},
		{"ErrInvalidData", ErrInvalidData, "invalid data"},
		{"ErrMaxRetriesExceeded", ErrMaxRetriesExceeded, "maximum retries exceeded"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.err, "%s is nil", tc.name)
			require.Equal(t, tc.expect, tc.err.Error(), "expected error message for %s", tc.name)
			require.True(t, errors.Is(tc.err, tc.err), "errors.Is failed for %s", tc.name)
		})
	}
}
