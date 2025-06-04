//go:build integration
// +build integration

// integration/grpc_integration_test.go
// Integration tests for gRPC transport using TransportX.
package integration

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/gozephyr/transportx"
	"github.com/gozephyr/transportx/protocols"
	grpcx "github.com/gozephyr/transportx/protocols/grpc"
)

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	addr := l.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

func TestGRPCTransportIntegration(t *testing.T) {
	port, err := getFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port: %v", err)
	}
	addr := "localhost:" + strconv.Itoa(port)

	t.Logf("[test] Starting server on %s", addr)
	ready := make(chan struct{})

	server := grpcx.NewServer(addr)
	go func() {
		close(ready)
		if err := server.Start(); err != nil {
			// Can't call t.Fatalf here, so print
			println("[test] Server failed to start:", err.Error())
		}
	}()
	<-ready
	time.Sleep(200 * time.Millisecond)
	defer server.Stop()

	cfg := grpcx.NewConfig()
	cfg.ServerAddress = "localhost"
	cfg.Port = port
	cfg.Timeout = 10 * time.Second
	cfg.MaxRetries = 2

	clientAny, err := transportx.NewClient(transportx.TypeGRPC, cfg)
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}
	client, ok := clientAny.(protocols.StreamProtocol)
	if !ok {
		t.Fatalf("Returned client does not implement protocols.StreamProtocol")
	}
	grpcTransport, ok := clientAny.(*grpcx.GRPCTransport)
	if !ok {
		t.Fatalf("Returned client is not *grpcx.GRPCTransport")
	}

	ctx := context.Background()
	t.Logf("[test] Connecting client")
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	t.Logf("[test] Client connected")
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("Failed to disconnect: %v", err)
		}
	}()

	t.Run("Single Request", func(t *testing.T) {
		data := []byte("Hello, gRPC Server!")
		if err := client.Send(ctx, data); err != nil {
			t.Errorf("Failed to send: %v", err)
		}
		resp, err := client.Receive(ctx)
		if err != nil {
			t.Errorf("Failed to receive: %v", err)
		}
		t.Logf("Received: %s", string(resp))
	})

	t.Run("Batch Request", func(t *testing.T) {
		batchData := [][]byte{
			[]byte("Batch 1"),
			[]byte("Batch 2"),
			[]byte("Batch 3"),
		}
		if err := grpcTransport.BatchSend(ctx, batchData); err != nil {
			t.Errorf("Failed to batch send: %v", err)
		}
		resp, err := grpcTransport.BatchReceive(ctx, len(batchData))
		if err != nil {
			t.Errorf("Failed to batch receive: %v", err)
		}
		t.Logf("Batch received: %v", resp)
	})

	t.Run("Metrics", func(t *testing.T) {
		metrics := grpcTransport.GetMetrics()
		if metrics == nil {
			t.Error("Expected non-nil metrics")
		}
	})
}
