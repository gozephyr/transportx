# TransportX API Overview

This document provides a high-level overview of the TransportX API, its main concepts, and usage patterns.

---

## 1. Architecture & Main Concepts

- **Unified Abstraction:** All protocols implement a common set of interfaces for connection, data transfer, and configuration.
- **Extensible Core:** New protocols can be added by implementing the relevant interfaces and registering them in the core package.
- **Configurable:** Each protocol has its own config struct with sensible defaults and builder methods.
- **Metrics & Observability:** Built-in support for metrics collection and export.

---

## 2. Key Interfaces

### `protocols.Client`

A unified interface for all protocol clients:

```go
type Client interface {
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    Send(ctx context.Context, data []byte) ([]byte, error)
    BatchSend(ctx context.Context, dataItems [][]byte) ([][]byte, error)
    GetMetrics() *metrics.Metrics
    IsConnected() bool
    SetTimeout(timeout time.Duration)
    GetTimeout() time.Duration
}
```

### `protocols.Protocol`

The base interface for all transports (see `protocols/protocol.go`).

### `protocols.StreamProtocol`, `BatchProtocol`, `PubSubProtocol`

Specialized interfaces for streaming, batch, and pub/sub protocols.

---

## 3. Core Package Functions

- `NewClient(typ TransportType, config protocols.Config) (protocols.Client, error)`
- `NewHTTPClient(config *http.Config) (protocols.Client, error)`
- `NewGRPCClient(config *grpc.Config) (protocols.Client, error)`
- `SupportedProtocols() []TransportType`
- `DefaultHTTPConfig() *http.Config`
- `DefaultGRPCConfig() *grpc.Config`

---

## 4. Example Usage Patterns

### Creating and Using a Client

```go
cfg := transportx.DefaultHTTPConfig()
cfg.ServerAddress = "localhost"
cfg.Port = 8080
client, err := transportx.NewClient(transportx.TypeHTTP, cfg)
if err != nil { /* handle error */ }
ctx := context.Background()
if err := client.Connect(ctx); err != nil { /* handle error */ }
resp, err := client.Send(ctx, []byte("Hello"))
```

### Listing Supported Protocols

```go
for _, proto := range transportx.SupportedProtocols() {
    fmt.Println("Supported:", proto)
}
```

---

## 5. Further Reference

- [Extension Guide](./EXTENSION_GUIDE.md)
- [GoDoc Reference](https://pkg.go.dev/github.com/gozephyr/transportx)

---
