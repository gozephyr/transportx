package protocols

import (
	"context"
	"time"

	"github.com/gozephyr/transportx/metrics"
)

// Client represents a generic client interface that can be used by all protocols
type Client interface {
	// Connect establishes a connection to the server
	Connect(ctx context.Context) error

	// Disconnect closes the connection
	Disconnect(ctx context.Context) error

	// Send sends a single request and returns the response
	Send(ctx context.Context, data []byte) ([]byte, error)

	// BatchSend sends multiple requests and returns the responses
	BatchSend(ctx context.Context, dataItems [][]byte) ([][]byte, error)

	// GetMetrics returns the client's metrics
	GetMetrics() *metrics.Metrics

	// IsConnected returns whether the client is connected
	IsConnected() bool

	// SetTimeout sets the client's timeout
	SetTimeout(timeout time.Duration)

	// GetTimeout returns the client's timeout
	GetTimeout() time.Duration
}
