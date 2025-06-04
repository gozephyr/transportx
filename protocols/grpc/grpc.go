// Package grpc implements a gRPC transport for the transportx library.
// It provides a client and server implementation for gRPC-based communication.
package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gozephyr/transportx/buffer"
	"github.com/gozephyr/transportx/metrics"
	generated "github.com/gozephyr/transportx/proto/generated/proto"
	"github.com/gozephyr/transportx/protocols"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// GRPCTransport implements the Protocol interface for gRPC
type GRPCTransport struct {
	config  *Config
	conn    *grpc.ClientConn
	buf     *buffer.Buffer
	state   protocols.ConnectionState
	metrics *metrics.Metrics
	mu      sync.RWMutex
	closed  bool
	client  generated.TransportServiceClient
}

// Ensure GRPCTransport implements protocols.StreamProtocol
var _ protocols.StreamProtocol = (*GRPCTransport)(nil)

// grpcStream wraps the generated gRPC stream and implements protocols.Stream
// Only basic Read/Write/Close/Context/Deadline methods are implemented for demo
// Real implementation should handle message marshaling and error handling

type grpcStream struct {
	stream grpc.BidiStreamingClient[generated.Message, generated.Message]
}

func (s *grpcStream) Read(p []byte) (int, error) {
	msg, err := s.stream.Recv()
	if err != nil {
		return 0, err
	}
	copy(p, msg.GetData())
	return len(msg.GetData()), nil
}

func (s *grpcStream) Write(p []byte) (int, error) {
	err := s.stream.Send(&generated.Message{Data: p})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *grpcStream) Close() error {
	return s.stream.CloseSend()
}

func (s *grpcStream) Context() context.Context {
	return nil // No context field anymore
}

func (s *grpcStream) SetDeadline(t time.Time) error      { return nil }
func (s *grpcStream) SetReadDeadline(t time.Time) error  { return nil }
func (s *grpcStream) SetWriteDeadline(t time.Time) error { return nil }

// NewGRPCTransport creates a new gRPC transport with optimized settings
func NewGRPCTransport(cfg *Config) (protocols.StreamProtocol, error) {
	if cfg == nil {
		cfg = NewConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Create metrics with detailed latency tracking
	metrics := metrics.NewMetrics(metrics.NewConfig().
		WithLatency(true).
		WithThroughput(true).
		WithErrors(true).
		WithRetries(true).
		WithConnections(true).
		WithActiveStreams(true).
		WithFailedConnects(true))

	// Create dial options
	dialOpts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cfg.KeepAliveTime,
			Timeout:             cfg.KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithInitialWindowSize(cfg.InitialWindowSize),
		grpc.WithMaxHeaderListSize(cfg.MaxHeaderListSize),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*10), // 10MB
			grpc.MaxCallSendMsgSize(1024*1024*10), // 10MB
		),
	}

	// Add TLS if configured
	if cfg.TLSConfig != nil {
		creds, err := loadTLSCredentials(cfg.TLSConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add custom dial options
	dialOpts = append(dialOpts, cfg.DialOptions...)

	// Create gRPC client using NewClient
	clientConn, err := grpc.Dial(cfg.GetFullAddress(), dialOpts...) //nolint:staticcheck // grpc.Dial is deprecated, but grpc.NewClient is not a drop-in replacement
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create gRPC client
	client := generated.NewTransportServiceClient(clientConn)

	return &GRPCTransport{
		config:  cfg,
		conn:    clientConn,
		client:  client,
		buf:     buffer.NewBuffer(),
		metrics: metrics,
		state: protocols.ConnectionState{
			Connected:    false,
			LastActivity: time.Now(),
			Metadata:     make(map[string]any),
		},
	}, nil
}

// GetProtocolName returns the name of the protocol
func (t *GRPCTransport) GetProtocolName() string {
	return "grpc"
}

// GetProtocolVersion returns the version of the protocol
func (t *GRPCTransport) GetProtocolVersion() string {
	return "1.0"
}

// SetTimeout sets the timeout for operations
func (t *GRPCTransport) SetTimeout(timeout time.Duration) {
	t.config.Timeout = timeout
}

// GetTimeout returns the current timeout
func (t *GRPCTransport) GetTimeout() time.Duration {
	return t.config.Timeout
}

// SetMaxRetries sets the maximum number of retries
func (t *GRPCTransport) SetMaxRetries(retries int) {
	t.config.MaxRetries = retries
}

// GetMaxRetries returns the maximum number of retries
func (t *GRPCTransport) GetMaxRetries() int {
	return t.config.MaxRetries
}

// IsConnected returns whether the transport is connected
func (t *GRPCTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state.Connected
}

// GetConnectionState returns the current connection state
func (t *GRPCTransport) GetConnectionState() protocols.ConnectionState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

// Connect establishes a connection to the gRPC server
func (t *GRPCTransport) Connect(ctx context.Context) error {
	if err := t.checkClosed(); err != nil {
		return err
	}

	t.mu.Lock()
	if t.state.Connected {
		t.mu.Unlock()
		return nil // Already connected
	}
	t.mu.Unlock()

	// Check connection state
	state := t.conn.GetState()
	if state == connectivity.Ready {
		t.updateState(true, nil)
		return nil
	}

	// Wait for connection to be ready
	for {
		if state == connectivity.Ready {
			t.updateState(true, nil)
			return nil
		}

		if !t.conn.WaitForStateChange(ctx, state) {
			t.updateState(false, ErrConnectionFailed)
			return ErrConnectionFailed
		}

		state = t.conn.GetState()
		if state == connectivity.Shutdown {
			t.updateState(false, ErrConnectionFailed)
			return ErrConnectionFailed
		}
	}
}

// Disconnect closes the gRPC connection
func (t *GRPCTransport) Disconnect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil // Already closed
	}

	t.closed = true
	t.state.Connected = false
	t.state.LastActivity = time.Now()

	if t.conn != nil {
		return t.conn.Close()
	}

	return nil
}

// Send sends data to the gRPC server
func (t *GRPCTransport) Send(ctx context.Context, data []byte) error {
	if err := t.checkClosed(); err != nil {
		return err
	}

	if !t.IsConnected() {
		if err := t.Connect(ctx); err != nil {
			return err
		}
	}

	// Send the message
	_, err := t.client.Send(ctx, &generated.Message{Data: data})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	t.updateState(true, nil)
	return nil
}

// Receive receives data from the gRPC server
func (t *GRPCTransport) Receive(ctx context.Context) ([]byte, error) {
	if err := t.checkClosed(); err != nil {
		return nil, err
	}

	if !t.IsConnected() {
		if err := t.Connect(ctx); err != nil {
			return nil, err
		}
	}

	// Receive the message
	msg, err := t.client.Receive(ctx, &generated.Empty{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReceiveFailed, err)
	}

	t.updateState(true, nil)
	return msg.GetData(), nil
}

// BatchSend sends multiple data items in a single request
func (t *GRPCTransport) BatchSend(ctx context.Context, dataItems [][]byte) error {
	if err := t.checkClosed(); err != nil {
		return err
	}

	if !t.IsConnected() {
		if err := t.Connect(ctx); err != nil {
			return err
		}
	}

	// Create batch message
	batch := &generated.BatchMessage{
		Messages: make([]*generated.Message, len(dataItems)),
	}
	for i, data := range dataItems {
		batch.GetMessages()[i] = &generated.Message{Data: data}
	}

	// Send batch
	_, err := t.client.BatchSend(ctx, batch)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	t.updateState(true, nil)
	return nil
}

// BatchReceive receives multiple data items in a single request
func (t *GRPCTransport) BatchReceive(ctx context.Context, batchSize int) ([][]byte, error) {
	if err := t.checkClosed(); err != nil {
		return nil, err
	}

	if !t.IsConnected() {
		if err := t.Connect(ctx); err != nil {
			return nil, err
		}
	}

	// Request batch
	batch, err := t.client.BatchReceive(ctx, &generated.BatchRequest{BatchSize: int32(batchSize)})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReceiveFailed, err)
	}

	// Convert messages to byte slices
	items := make([][]byte, len(batch.GetMessages()))
	for i, msg := range batch.GetMessages() {
		items[i] = msg.GetData()
	}

	t.updateState(true, nil)
	return items, nil
}

// GetMetrics returns the transport metrics
func (t *GRPCTransport) GetMetrics() *metrics.Metrics {
	return t.metrics
}

// AcceptStream is not implemented for client transport
func (t *GRPCTransport) AcceptStream(ctx context.Context) (protocols.Stream, error) {
	return nil, errors.New("AcceptStream not implemented on client transport")
}

// NewStream implements protocols.StreamProtocol for GRPCTransport
func (t *GRPCTransport) NewStream(ctx context.Context) (protocols.Stream, error) {
	if err := t.checkClosed(); err != nil {
		return nil, err
	}
	if !t.IsConnected() {
		if err := t.Connect(ctx); err != nil {
			return nil, err
		}
	}
	stream, err := t.client.StreamMessages(ctx)
	if err != nil {
		return nil, err
	}
	return &grpcStream{stream: stream}, nil
}

// updateState updates the connection state
func (t *GRPCTransport) updateState(connected bool, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.state.Connected = connected
	t.state.LastActivity = time.Now()
	t.state.Error = err
}

// checkClosed checks if the transport is closed
func (t *GRPCTransport) checkClosed() error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.closed {
		return errors.New("transport is closed")
	}
	return nil
}

// loadTLSCredentials loads TLS credentials from the configuration
func loadTLSCredentials(cfg *TLSConfig) (credentials.TransportCredentials, error) {
	if cfg.InsecureSkipVerify {
		return credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		}), nil
	}

	certificate, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("could not load client key pair: %w", err)
	}

	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(cfg.CAFile)
	if err != nil {
		return nil, fmt.Errorf("could not read CA certificate: %w", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, errors.New("failed to append CA certificate")
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
		ServerName:   cfg.ServerName,
	}), nil
}
