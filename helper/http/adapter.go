// Package http provides HTTP helpers for TransportX, including adapters and roundtrippers.
package http

import (
	"context"
	"io"
	"net/http"
)

// HandlerFunc is the signature for business logic handlers compatible with the adapter.
//
// This allows you to write your business logic as a simple function that takes a context and request body,
// and returns a response and error. This function can be reused across different frameworks and is easy to test.
//
// Example:
//
//	func MyLogic(ctx context.Context, data []byte) ([]byte, error) {
//	    // process data
//	    return []byte("hello world"), nil
//	}
type HandlerFunc func(ctx context.Context, data []byte) ([]byte, error)

// HTTPAdapter returns a net/http compatible handler that wraps the given business logic handler.
// It reads the request body, calls the handler, and writes the response or error.
//
// Usage Example:
//
//	import (
//	    "net/http"
//	    helperhttp "github.com/gozephyr/transportx/helper/http"
//	)
//
//	func MyLogic(ctx context.Context, data []byte) ([]byte, error) {
//	    // Your business logic here
//	    return []byte("Hello from server!"), nil
//	}
//
//	func main() {
//	    http.HandleFunc("/api", helperhttp.HTTPAdapter(MyLogic))
//	    http.ListenAndServe(":8080", nil)
//	}
//
// Benefits:
//   - Decouples business logic from HTTP details
//   - Reduces boilerplate for request/response handling
//   - Makes business logic easy to test and reuse
//   - Foundation for adapters for other frameworks (Gin, Fiber, etc.)
func HTTPAdapter(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		resp, err := handler(r.Context(), data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write(resp); writeErr != nil {
			// Optionally log the error or write a generic error
			http.Error(w, "failed to write response", http.StatusInternalServerError)
		}
	}
}
