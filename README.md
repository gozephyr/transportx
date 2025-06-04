# TransportX 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/gozephyr/transportx.svg)](https://pkg.go.dev/github.com/gozephyr/transportx)
[![Go Report Card](https://goreportcard.com/badge/github.com/gozephyr/transportx)](https://goreportcard.com/report/github.com/gozephyr/transportx)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Build Status](https://github.com/gozephyr/transportx/actions/workflows/go.yml/badge.svg)](https://github.com/gozephyr/transportx/actions/workflows/go.yml)
[![Coverage Status](https://img.shields.io/codecov/c/github/gozephyr/transportx?label=coverage)](https://app.codecov.io/gh/gozephyr/transportx)

**TransportX** is a high-performance, extensible transport layer library for Go. It provides a unified, user-friendly interface for multiple transport protocols (HTTP, gRPC, and more), designed to be faster and more efficient than the standard Go transport packages while maintaining simplicity and ease of use.

---

## ✨ Features

- 🚀 **Unified API** for HTTP, gRPC, and future protocols
- 🧩 **Extensible**: Add your own protocols easily
- 📊 **Built-in metrics** and observability
- 🛡️ **Production-ready**: Robust connection pooling, DNS, and error handling
- 🔒 **Security**: TLS and authentication support (where applicable)
- 🧪 **Tested**: Includes examples and stress tests

---

## 💡 Why Use TransportX?

TransportX is designed for developers who want a robust, flexible, and high-performance way to handle network communication in Go. It abstracts away the complexity of managing different transport protocols, connection pooling, retries, and observability, letting you focus on your business logic. Whether you're building microservices, distributed systems, or high-throughput applications, TransportX provides a consistent and extensible API for all your transport needs.

### 🌟 Key Benefits

- 🔄 **Unified API:** Switch between HTTP, gRPC, and other protocols with minimal code changes.
- 🔁 **Advanced Connection Management:** Built-in pooling, retries, timeouts, and health checks.
- ⚡ **Performance:** Optimized for throughput and latency, with support for HTTP/2 and efficient batching.
- 📈 **Observability:** Built-in metrics for latency, throughput, errors, and connection state.
- 🧱 **Extensibility:** Easily add support for new protocols.
- 🧹 **Cleaner Code:** Focus on your application logic, not low-level networking details.

---

## 📊 Summary Comparison Table

| 🚦 Feature/Benefit         | 🌐 net/http (std) | 🔗 gRPC (std) | 🚀 TransportX |
|---------------------------|:----------------:|:------------:|:------------:|
| Unified API (multi-proto) |        ❌         |      ❌      |      ✅      |
| Connection Pooling        | Basic/Manual ⚙️  |  Built-in 🏗️ |  Advanced 🚀 |
| Retries/Timeouts          | Manual/Basic ⏱️  |   Basic ⏱️   |  Advanced 🔁 |
| Metrics/Observability     |        ❌         |  Partial 📉  |      ✅      |
| Protocol Extensibility    |        ❌         |      ❌      |      ✅      |
| Batching Support          |        ❌         |  Partial 📦  |      ✅      |
| Health Checks             |        ❌         |  Partial 🩺  |      ✅      |
| Easy Config Management    |     Manual 📝     |  Manual 📝   |      ✅      |
| Future Protocols (WebSocket, QUIC, etc.) | ❌ | ❌ | ✅ (planned) |

**Legend:**

- ✅ = Fully supported or built-in
- ❌ = Not supported or requires significant manual work
- Partial = Some support, but not unified or requires extra setup
- Emojis indicate special features or manual effort

---

## 📦 Installation

```bash
go get github.com/gozephyr/transportx
```

---

## 🚀 Quick Start

### 🌍 HTTP Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/gozephyr/transportx"
)

func main() {
    // Create a default HTTP config and customize as needed
    cfg := transportx.DefaultHTTPConfig()
    cfg.ServerAddress = "localhost"
    cfg.Port = 8080

    // Create a unified client
    client, err := transportx.NewClient(transportx.TypeHTTP, cfg)
    if err != nil {
        panic(err)
    }
    defer client.Disconnect(context.Background())

    // Connect and send a request
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        panic(err)
    }
    resp, err := client.Send(ctx, []byte("Hello, server!"))
    if err != nil {
        panic(err)
    }
    fmt.Println("Response:", string(resp))
}
```

### ⚡ gRPC Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/gozephyr/transportx"
)

func main() {
    cfg := transportx.DefaultGRPCConfig()
    cfg.ServerAddress = "localhost"
    cfg.Port = 50051

    client, err := transportx.NewClient(transportx.TypeGRPC, cfg)
    if err != nil {
        panic(err)
    }
    defer client.Disconnect(context.Background())

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        panic(err)
    }
    resp, err := client.Send(ctx, []byte("Hello, gRPC server!"))
    if err != nil {
        panic(err)
    }
    fmt.Println("gRPC Response:", string(resp))
}
```

---

## ⚙️ Configuration Options

### 🌍 HTTP Config (`DefaultHTTPConfig()`)

| Field               | Type           | Default           | Description                                 |
|---------------------|----------------|-------------------|---------------------------------------------|
| Protocol            | string         | "http"           | Protocol scheme ("http" or "https")         |
| ServerAddress       | string         | "localhost"      | Hostname or IP                              |
| Port                | int            | 8080              | Server port                                 |
| MaxIdleConns        | int            | 100               | Max idle connections                        |
| MaxConnsPerHost     | int            | 100               | Max connections per host                    |
| IdleTimeout         | time.Duration  | 90s               | Idle connection timeout                     |
| KeepAlive           | time.Duration  | 30s               | Keep-alive duration                         |
| ResponseTimeout     | time.Duration  | 30s               | Response timeout                            |
| MaxWaitDuration     | time.Duration  | 5s                | Max wait for connection                     |
| HealthCheckDelay    | time.Duration  | 30s               | Health check delay                          |
| EnableHTTP2         | bool           | true              | Enable HTTP/2                               |
| MaxHeaderBytes      | int            | 32*1024           | Max header bytes                            |
| DisableCompression  | bool           | false             | Disable compression                         |
| DisableKeepAlives   | bool           | false             | Disable keep-alives                         |

### ⚡ gRPC Config (`DefaultGRPCConfig()`)

| Field                 | Type           | Default           | Description                                 |
|-----------------------|----------------|-------------------|---------------------------------------------|
| ServerAddress         | string         | "localhost"      | Hostname or IP                              |
| Port                  | int            | 50051             | Server port                                 |
| Timeout               | time.Duration  | 30s               | Operation timeout                           |
| MaxRetries            | int            | 3                 | Max retries                                 |
| MaxConcurrentStreams  | uint32         | 100               | Max concurrent streams                      |
| InitialWindowSize     | int32          | 1MB               | Initial window size                         |
| MaxHeaderListSize     | uint32         | 8192              | Max header list size                        |
| KeepAliveTime         | time.Duration  | 30s               | Keep-alive ping interval                    |
| KeepAliveTimeout      | time.Duration  | 10s               | Keep-alive timeout                          |
| TLSConfig             | *TLSConfig     | nil               | TLS configuration (see below)               |
| DialOptions           | []DialOption   | nil               | Additional gRPC dial options                |

---

## 🧭 Protocol Support Matrix

| Protocol   | Status      | Notes                        |
|------------|-------------|------------------------------|
| 🌍 HTTP       | ✅ Stable   | Full client support          |
| ⚡ gRPC       | ✅ Stable   | Full client support          |
| 🔌 TCP        | 🚧 Planned  | Not yet implemented          |
| 📡 UDP        | 🚧 Planned  | Not yet implemented          |
| 🌐 WebSocket  | 🚧 Planned  | Not yet implemented          |
| 🏎️ QUIC       | 🚧 Planned  | Not yet implemented          |
| 📻 MQTT       | 🚧 Planned  | Not yet implemented          |
| ✉️ AMQP       | 🚧 Planned  | Not yet implemented          |

---

## 🛠️ Extending TransportX

TransportX is designed for extensibility. To add a new protocol:

1. Implement the relevant interface(s) from `protocols/` (e.g., `Protocol`, `StreamProtocol`, `BatchProtocol`, `PubSubProtocol`).
2. Provide a config struct and builder.
3. Register your protocol in the core package (see `transportx.go`).
4. Add tests and examples.

---

## 📚 Further Documentation

- [GoDoc Reference](https://pkg.go.dev/github.com/gozephyr/transportx)
<!-- - [Examples](./examples/) -->
- [API & Extension Guide](./docs/)

---

## 📝 License

TransportX is licensed under the MIT License.
