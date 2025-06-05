//go:build integration
// +build integration

// integration/http_integration_test.go
// Integration tests for HTTP transport using TransportX.
package integration

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gozephyr/transportx"
	helperhttp "github.com/gozephyr/transportx/helper/http"
	"github.com/gozephyr/transportx/protocols"
	httpx "github.com/gozephyr/transportx/protocols/http"
)

// Test constants
const (
	testPort        = 8081
	maxIdleConns    = 2000
	maxConnsPerHost = 2000
)

func TestHTTPTransportIntegration(t *testing.T) {
	server := startExampleServer(t)
	defer server.Close()

	cfg := transportx.DefaultHTTPConfig()
	cfg.ServerAddress = "localhost"
	cfg.Port = testPort
	cfg.Timeout = 30 * time.Second
	cfg.MaxRetries = 3
	cfg.MaxIdleConns = maxIdleConns
	cfg.MaxConnsPerHost = maxConnsPerHost
	cfg.IdleTimeout = 90 * time.Second
	cfg.KeepAlive = 30 * time.Second
	cfg.ResponseTimeout = 30 * time.Second
	cfg.MaxWaitDuration = 10 * time.Second
	cfg.HealthCheckDelay = 30 * time.Second
	cfg.EnableHTTP2 = true
	cfg.DisableKeepAlives = false

	clientAny, err := transportx.NewClient(transportx.TypeHTTP, cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client, ok := clientAny.(protocols.Client)
	if !ok {
		t.Fatalf("Returned client does not implement protocols.Client")
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("Failed to disconnect: %v", err)
		}
	}()

	t.Run("Single Request", func(t *testing.T) {
		data := []byte("Hello, Server!")
		resp, err := client.Send(ctx, data)
		if err != nil {
			t.Errorf("Failed to send request: %v", err)
		}
		if len(resp) == 0 {
			t.Errorf("Expected non-empty response")
		}
	})

	t.Run("Batch Request", func(t *testing.T) {
		batchData := [][]byte{
			[]byte("Batch item 1"),
			[]byte("Batch item 2"),
			[]byte("Batch item 3"),
		}
		batchResp, err := client.BatchSend(ctx, batchData)
		if err != nil {
			t.Errorf("Failed to send batch request: %v", err)
		}
		if len(batchResp) != len(batchData) {
			t.Errorf("Expected batch response of size %d, got %d", len(batchData), len(batchResp))
		}
	})

	t.Run("Metrics", func(t *testing.T) {
		metrics := client.GetMetrics()
		if metrics == nil {
			t.Error("Expected non-nil metrics")
		}
	})
}

func startExampleServer(t *testing.T) *http.Server {
	cfg := httpx.NewConfig()
	cfg.ServerAddress = "localhost"
	cfg.Port = testPort
	cfg.Timeout = 30 * time.Second
	cfg.MaxRetries = 3
	cfg.MaxIdleConns = maxIdleConns
	cfg.MaxConnsPerHost = maxConnsPerHost
	cfg.IdleTimeout = 90 * time.Second
	cfg.KeepAlive = 30 * time.Second
	cfg.ResponseTimeout = 30 * time.Second
	cfg.MaxWaitDuration = 10 * time.Second
	cfg.HealthCheckDelay = 30 * time.Second
	cfg.EnableHTTP2 = true
	cfg.DisableKeepAlives = false

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Received request: %s %s", r.Method, r.URL.Path)
		batchSize := r.Header.Get("X-Batch-Size")
		if batchSize == "" {
			batchSize = "1"
		}
		size, err := strconv.Atoi(batchSize)
		if err != nil {
			t.Logf("Invalid batch size: %v", err)
			http.Error(w, "Invalid batch size", http.StatusBadRequest)
			return
		}
		w.Header().Set("X-Batch-Size", batchSize)

		switch r.Method {
		case http.MethodPost:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Logf("Failed to read request body: %v", err)
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			offset := 0
			var batchData []byte
			for i := 0; i < size && offset+4 <= len(body); i++ {
				if offset+4 > len(body) {
					break
				}
				itemLen := int(binary.BigEndian.Uint32(body[offset : offset+4]))
				offset += 4
				if offset+itemLen > len(body) {
					break
				}
				item := body[offset : offset+itemLen]
				offset += itemLen
				sizeBytes := make([]byte, 4)
				binary.BigEndian.PutUint32(sizeBytes, uint32(len(item)))
				batchData = append(batchData, sizeBytes...)
				batchData = append(batchData, item...)
			}
			if _, err := w.Write(batchData); err != nil {
				t.Logf("Failed to write response: %v", err)
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
		case http.MethodGet:
			var batchData []byte
			for i := 0; i < size; i++ {
				response := fmt.Sprintf("Batch response %d", i+1)
				responseBytes := []byte(response)
				sizeBytes := make([]byte, 4)
				binary.BigEndian.PutUint32(sizeBytes, uint32(len(responseBytes)))
				batchData = append(batchData, sizeBytes...)
				batchData = append(batchData, responseBytes...)
			}
			if _, err := w.Write(batchData); err != nil {
				t.Logf("Failed to write response: %v", err)
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
		}
	})

	mux.HandleFunc("/batch", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Received batch request: %s %s", r.Method, r.URL.Path)
		batchSize := r.Header.Get("X-Batch-Size")
		if batchSize == "" {
			batchSize = "1"
		}
		t.Logf("Batch size: %s", batchSize)

		size, err := strconv.Atoi(batchSize)
		if err != nil {
			t.Logf("Invalid batch size: %v", err)
			http.Error(w, "Invalid batch size", http.StatusBadRequest)
			return
		}

		w.Header().Set("X-Batch-Size", batchSize)

		switch r.Method {
		case http.MethodPost:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Logf("Failed to read request body: %v", err)
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			t.Logf("Received batch POST body: %s", string(body))

			offset := 0
			var batchData []byte
			for i := 0; i < size && offset+4 <= len(body); i++ {
				if offset+4 > len(body) {
					break
				}
				itemLen := int(binary.BigEndian.Uint32(body[offset : offset+4]))
				offset += 4
				if offset+itemLen > len(body) {
					break
				}
				item := body[offset : offset+itemLen]
				offset += itemLen

				sizeBytes := make([]byte, 4)
				binary.BigEndian.PutUint32(sizeBytes, uint32(len(item)))
				batchData = append(batchData, sizeBytes...)
				batchData = append(batchData, item...)
			}
			if _, err := w.Write(batchData); err != nil {
				t.Logf("Failed to write response: %v", err)
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
		case http.MethodGet:
			var batchData []byte
			for i := 0; i < size; i++ {
				response := fmt.Sprintf("Batch response %d", i+1)
				responseBytes := []byte(response)

				sizeBytes := make([]byte, 4)
				binary.BigEndian.PutUint32(sizeBytes, uint32(len(responseBytes)))
				batchData = append(batchData, sizeBytes...)
				batchData = append(batchData, responseBytes...)
			}
			if _, err := w.Write(batchData); err != nil {
				t.Logf("Failed to write response: %v", err)
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				return
			}
		}
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.ResponseTimeout,
		WriteTimeout: cfg.ResponseTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		t.Logf("Starting server on port %d...", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server failed: %v", err)
		}
	}()

	t.Log("Waiting for server to start...")
	time.Sleep(500 * time.Millisecond)
	t.Log("Server started")

	return server
}

func TestHTTPAdapter_Integration(t *testing.T) {
	// Start a test server using HTTPAdapter
	handler := func(ctx context.Context, data []byte) ([]byte, error) {
		return append([]byte("echo: "), data...), nil
	}
	ts := http.Server{
		Addr:    ":18081",
		Handler: http.HandlerFunc(helperhttp.HTTPAdapter(handler)),
	}
	go ts.ListenAndServe()
	defer ts.Close()
	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Post("http://localhost:18081", "text/plain", strings.NewReader("integration"))
	if err != nil {
		t.Fatalf("Failed to POST: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	if string(body) != "echo: integration" {
		t.Errorf("Unexpected response: %s", body)
	}
}

type proxyTransport struct{}

func (p *proxyTransport) Send(ctx context.Context, data []byte) error { return nil }
func (p *proxyTransport) Receive(ctx context.Context) ([]byte, error) {
	return []byte("rt: integration"), nil
}

func TestTransportxRoundTripper_Integration(t *testing.T) {
	// Start a test server that echoes the request body
	ts := http.Server{
		Addr: ":18082",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, _ := io.ReadAll(r.Body)
			w.Write(append([]byte("rt: "), data...))
		}),
	}
	go ts.ListenAndServe()
	defer ts.Close()
	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	rt := helperhttp.NewTransportxRoundTripper(&proxyTransport{})
	client := &http.Client{Transport: rt}
	resp, err := client.Post("http://localhost:18082", "text/plain", strings.NewReader("integration"))
	if err != nil {
		t.Fatalf("Failed to POST: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	if string(body) != "rt: integration" {
		t.Errorf("Unexpected response: %s", body)
	}
}
