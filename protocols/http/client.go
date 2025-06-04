package http

import (
	"context"
	"time"

	"github.com/gozephyr/transportx/metrics"
	"github.com/gozephyr/transportx/protocols"
)

// Client represents an HTTP client with a clean API
type Client struct {
	transport *HTTPTransport
	config    *Config
}

// NewClient creates a new HTTP client with the given configuration
func NewClient(cfg *Config) (protocols.Client, error) {
	if cfg == nil {
		cfg = NewConfig()
	}

	transport, err := NewHTTPTransport(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		transport: transport.(*HTTPTransport),
		config:    cfg,
	}, nil
}

// Connect establishes a connection to the server
func (c *Client) Connect(ctx context.Context) error {
	return c.transport.Connect(ctx)
}

// Disconnect closes the connection
func (c *Client) Disconnect(ctx context.Context) error {
	return c.transport.Disconnect(ctx)
}

// Send sends a single request and returns the response
func (c *Client) Send(ctx context.Context, data []byte) ([]byte, error) {
	if err := c.transport.Send(ctx, data); err != nil {
		return nil, err
	}
	return c.transport.Receive(ctx)
}

// BatchSend sends multiple requests and returns the responses
func (c *Client) BatchSend(ctx context.Context, dataItems [][]byte) ([][]byte, error) {
	if err := c.transport.BatchSend(ctx, dataItems); err != nil {
		return nil, err
	}
	return c.transport.BatchReceive(ctx, len(dataItems))
}

// GetMetrics returns the client's metrics
func (c *Client) GetMetrics() *metrics.Metrics {
	return c.transport.GetMetrics()
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.transport.IsConnected()
}

// SetTimeout sets the client's timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.transport.SetTimeout(timeout)
}

// GetTimeout returns the client's timeout
func (c *Client) GetTimeout() time.Duration {
	return c.transport.GetTimeout()
}
