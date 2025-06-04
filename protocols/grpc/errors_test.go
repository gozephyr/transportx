package grpc

import (
	"errors"
	"testing"
)

func TestErrorVariables(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		expect string
	}{
		{"ErrInvalidServerAddress", ErrInvalidServerAddress, "invalid server address"},
		{"ErrInvalidPort", ErrInvalidPort, "invalid port number"},
		{"ErrInvalidTimeout", ErrInvalidTimeout, "invalid timeout"},
		{"ErrInvalidMaxRetries", ErrInvalidMaxRetries, "invalid max retries"},
		{"ErrInvalidMaxConcurrentStreams", ErrInvalidMaxConcurrentStreams, "invalid max concurrent streams"},
		{"ErrInvalidInitialWindowSize", ErrInvalidInitialWindowSize, "invalid initial window size"},
		{"ErrInvalidMaxHeaderListSize", ErrInvalidMaxHeaderListSize, "invalid max header list size"},
		{"ErrInvalidKeepAliveTime", ErrInvalidKeepAliveTime, "invalid keepalive time"},
		{"ErrInvalidKeepAliveTimeout", ErrInvalidKeepAliveTimeout, "invalid keepalive timeout"},
		{"ErrConnectionFailed", ErrConnectionFailed, "connection failed"},
		{"ErrSendFailed", ErrSendFailed, "send failed"},
		{"ErrReceiveFailed", ErrReceiveFailed, "receive failed"},
		{"ErrStreamClosed", ErrStreamClosed, "stream closed"},
		{"ErrInvalidMessage", ErrInvalidMessage, "invalid message"},
		{"ErrInvalidStreamState", ErrInvalidStreamState, "invalid stream state"},
		{"ErrStreamTimeout", ErrStreamTimeout, "stream timeout"},
		{"ErrStreamCancelled", ErrStreamCancelled, "stream cancelled"},
		{"ErrStreamReset", ErrStreamReset, "stream reset"},
		{"ErrStreamUnavailable", ErrStreamUnavailable, "stream unavailable"},
		{"ErrStreamExhausted", ErrStreamExhausted, "stream exhausted"},
		{"ErrStreamLimitExceeded", ErrStreamLimitExceeded, "stream limit exceeded"},
		{"ErrStreamDeadlineExceeded", ErrStreamDeadlineExceeded, "stream deadline exceeded"},
		{"ErrStreamResourceExhausted", ErrStreamResourceExhausted, "stream resource exhausted"},
		{"ErrStreamUnimplemented", ErrStreamUnimplemented, "stream unimplemented"},
		{"ErrStreamInternal", ErrStreamInternal, "stream internal error"},
		{"ErrStreamDataLoss", ErrStreamDataLoss, "stream data loss"},
		{"ErrStreamFailedPrecondition", ErrStreamFailedPrecondition, "stream failed precondition"},
		{"ErrStreamAborted", ErrStreamAborted, "stream aborted"},
		{"ErrStreamOutOfRange", ErrStreamOutOfRange, "stream out of range"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatalf("%s is nil", tc.name)
			}
			if tc.err.Error() != tc.expect {
				t.Errorf("expected error message '%s', got '%s'", tc.expect, tc.err.Error())
			}
			if !errors.Is(tc.err, tc.err) {
				t.Errorf("errors.Is failed for %s", tc.name)
			}
		})
	}
}
