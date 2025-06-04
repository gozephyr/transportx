# TransportX Extension Guide

TransportX is designed to be highly extensible. This guide will help you add new transport protocols to the library, whether for your own use or to contribute upstream.

---

## 1. Overview of Extensibility

TransportX uses Go interfaces to define protocol behaviors. You can add new protocols (e.g., TCP, WebSocket, custom binary) by implementing the relevant interfaces and registering your protocol in the core package.

---

## 2. Steps to Add a New Protocol

### Step 1: Implement the Protocol Interface

Choose the interface(s) that match your protocol's capabilities:

- `Protocol` (basic connect/send/receive)
- `StreamProtocol` (for streaming)
- `BatchProtocol` (for batch operations)
- `PubSubProtocol` (for publish/subscribe)

Example:

```go
// myproto.go
package myproto

type MyProtoTransport struct { /* ... */ }

func (t *MyProtoTransport) Connect(ctx context.Context) error { /* ... */ }
func (t *MyProtoTransport) Disconnect(ctx context.Context) error { /* ... */ }
func (t *MyProtoTransport) Send(ctx context.Context, data []byte) error { /* ... */ }
func (t *MyProtoTransport) Receive(ctx context.Context) ([]byte, error) { /* ... */ }
// ... implement other required methods ...
```

### Step 2: Create a Config Struct and Builder

Provide a configuration struct and a builder function for your protocol:

```go
type Config struct {
    // Protocol-specific options
}

func NewConfig() *Config {
    return &Config{ /* sensible defaults */ }
}
```

### Step 3: Register Your Protocol in the Core Package

Add a factory function in the core (e.g., `transportx.go`) to create your protocol:

```go
func NewMyProtoClient(config *myproto.Config) (protocols.Client, error) {
    return myproto.NewClient(config)
}
```

And update the protocol switch in `NewClient` to support your new type.

### Step 4: Add Tests and Examples

- Add unit tests for your protocol implementation.
- Add an example in the `examples/` directory.

---

## 3. Best Practices

- Follow the patterns of existing protocols (HTTP, gRPC) for consistency.
- Use sensible defaults in your config builder.
- Document all public types and methods with GoDoc comments.
- Ensure thread safety for shared state.
- Add metrics support if possible.

---

## 4. Contributing Extensions Upstream

- Fork the repo and create a feature branch.
- Add your protocol, tests, and documentation.
- Open a pull request with a clear description.
- Follow the project's code style and contribution guidelines.

---

## 5. Further Reference

- [API Overview](./API_OVERVIEW.md)
- [GoDoc Reference](https://pkg.go.dev/github.com/gozephyr/transportx)

---
