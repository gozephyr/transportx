package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// minimalTransport is the minimal interface required for TransportxRoundTripper.
type minimalTransport interface {
	Send(ctx context.Context, data []byte) error
	Receive(ctx context.Context) ([]byte, error)
}

// TransportxRoundTripper implements http.RoundTripper using a minimal transport interface.
//
// This allows you to use your custom transportx logic as a drop-in replacement for the standard
// net/http transport in any http.Client. This is useful for adding custom protocol handling,
// metrics, retries, or other features provided by transportx, while keeping compatibility with
// the Go standard library HTTP client interface.
//
// Usage Example:
//
//	import (
//	    "net/http"
//	    txhttp "github.com/gozephyr/transportx/protocols/http"
//	    helperhttp "github.com/gozephyr/transportx/helper/http"
//	)
//
//	func main() {
//	    // 1. Create a transportx HTTPTransport (as client)
//	    cfg := txhttp.NewConfig().WithServerAddress("example.com").WithPort(8080)
//	    transport, err := txhttp.NewHTTPTransport(cfg)
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    // 2. Wrap it as a RoundTripper
//	    rt := helperhttp.NewTransportxRoundTripper(transport.(*txhttp.HTTPTransport))
//
//	    // 3. Use it in a standard http.Client
//	    client := &http.Client{
//	        Transport: rt,
//	    }
//
//	    // 4. Use client as usual
//	    resp, err := client.Post("http://example.com", "application/json", bytes.NewReader([]byte(`{"foo":"bar"}`)))
//	    // ... handle resp and err ...
//	}
//
// Benefits:
//   - Plug-and-play with any code using http.Client
//   - Enables custom transportx features (metrics, batching, etc.)
//   - Easy to test and extend
type TransportxRoundTripper struct {
	transport minimalTransport
}

// NewTransportxRoundTripper creates a new TransportxRoundTripper with the given transportx HTTPTransport.
func NewTransportxRoundTripper(transport minimalTransport) *TransportxRoundTripper {
	return &TransportxRoundTripper{transport: transport}
}

// RoundTrip executes a single HTTP transaction using the minimal transport interface.
// It reads the request body, sends it using transportx, receives the response, and returns it as an http.Response.
func (t *TransportxRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	var body []byte
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		body = b
	}
	// Send the request body using transportx
	if err := t.transport.Send(ctx, body); err != nil {
		return nil, err
	}
	// Receive the response using transportx
	respBody, err := t.transport.Receive(ctx)
	if err != nil {
		return nil, err
	}
	// Create a new http.Response
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(respBody)),
		Header:     make(http.Header),
		Request:    req,
	}
	return resp, nil
}
