// Package transportx provides the main factory and types for transportx.
package transportx

import (
	"errors"

	txerrors "github.com/gozephyr/transportx/errors"
	"github.com/gozephyr/transportx/protocols"
	"github.com/gozephyr/transportx/protocols/grpc"
	"github.com/gozephyr/transportx/protocols/http"
)

// TransportType represents the type of transport to create
type TransportType string

const (
	// TypeHTTP is a standard HTTP transport
	TypeHTTP TransportType = "http"

	// TypeGRPC is a gRPC transport
	TypeGRPC TransportType = "grpc"
)

// Transportx creates and configures transports
type Transportx struct {
	// Transport type
	transportType TransportType
	// Configuration
	config protocols.Config
}

// NewTransportx creates a new transportx instance
func NewTransportx(transportType TransportType, cfg protocols.Config) *Transportx {
	return &Transportx{
		config:        cfg,
		transportType: transportType,
	}
}

// WithType sets the transport type
func (t *Transportx) WithType(transportType TransportType) *Transportx {
	t.transportType = transportType
	return t
}

// WithConfig sets the transport configuration
func (t *Transportx) WithConfig(config protocols.Config) *Transportx {
	if config != nil {
		t.config = config
	}
	return t
}

// CreateHTTP creates a new HTTP transport with the given configuration
func (t *Transportx) CreateHTTP(cfg *http.Config) (protocols.Protocol, error) {
	if cfg == nil {
		cfg = http.NewConfig()
	}
	return http.NewHTTPTransport(cfg)
}

// CreateGRPC creates a new gRPC transport
func (t *Transportx) CreateGRPC(cfg *grpc.Config) (protocols.StreamProtocol, error) {
	if cfg == nil {
		cfg = grpc.NewConfig()
	}
	return grpc.NewGRPCTransport(cfg)
}

// CreateTCP creates a new TCP transport
func (t *Transportx) CreateTCP() (protocols.Protocol, error) {
	return nil, txerrors.ErrUnsupportedTransport // TODO: Implement TCP transport
}

// CreateUDP creates a new UDP transport
func (t *Transportx) CreateUDP() (protocols.Protocol, error) {
	return nil, txerrors.ErrUnsupportedTransport // TODO: Implement UDP transport
}

// CreateWebSocket creates a new WebSocket transport
func (t *Transportx) CreateWebSocket() (protocols.StreamProtocol, error) {
	return nil, txerrors.ErrUnsupportedTransport // TODO: Implement WebSocket transport
}

// CreateQUIC creates a new QUIC transport
func (t *Transportx) CreateQUIC() (protocols.StreamProtocol, error) {
	return nil, txerrors.ErrUnsupportedTransport // TODO: Implement QUIC transport
}

// CreateMQTT creates a new MQTT transport
func (t *Transportx) CreateMQTT() (protocols.PubSubProtocol, error) {
	return nil, txerrors.ErrUnsupportedTransport // TODO: Implement MQTT transport
}

// CreateAMQP creates a new AMQP transport
func (t *Transportx) CreateAMQP() (protocols.PubSubProtocol, error) {
	return nil, txerrors.ErrUnsupportedTransport // TODO: Implement AMQP transport
}

// NewClient creates a new client for the given transport type and config.
// For HTTP, returns protocols.Client. For gRPC, returns protocols.StreamProtocol.
func NewClient(typ TransportType, config protocols.Config) (any, error) {
	// Switch on type and delegate to protocol-specific helpers
	switch typ {
	case TypeHTTP:
		httpCfg, ok := config.(*http.Config)
		if !ok {
			return nil, errors.New("config must be of type *http.Config for HTTP client")
		}
		return http.NewClient(httpCfg)
	case TypeGRPC:
		grpcCfg, ok := config.(*grpc.Config)
		if !ok {
			return nil, errors.New("config must be of type *grpc.Config for gRPC client")
		}
		// Returns protocols.StreamProtocol for gRPC
		return grpc.NewGRPCTransport(grpcCfg)
	default:
		return nil, txerrors.ErrUnsupportedTransport
	}
}

// NewHTTPClient creates a new HTTP client with the given config.
func NewHTTPClient(config *http.Config) (protocols.Client, error) {
	return http.NewClient(config)
}

// NewGRPCClient creates a new gRPC client with the given config.
// Returns protocols.StreamProtocol for gRPC.
func NewGRPCClient(config *grpc.Config) (protocols.StreamProtocol, error) {
	return grpc.NewGRPCTransport(config)
}

// SupportedProtocols returns a list of supported transport types.
func SupportedProtocols() []TransportType {
	return []TransportType{TypeHTTP, TypeGRPC}
}

// DefaultHTTPConfig returns a new HTTP config with sensible defaults.
func DefaultHTTPConfig() *http.Config {
	return http.NewConfig()
}

// DefaultGRPCConfig returns a new gRPC config with sensible defaults.
func DefaultGRPCConfig() *grpc.Config {
	return grpc.NewConfig()
}
