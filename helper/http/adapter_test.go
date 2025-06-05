package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPAdapter_Success(t *testing.T) {
	handler := HTTPAdapter(func(ctx context.Context, data []byte) ([]byte, error) {
		require.Equal(t, "input", string(data))
		return []byte("output"), nil
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader("input"))
	rw := httptest.NewRecorder()

	handler(rw, req)

	resp := rw.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := rw.Body.String()
	require.Equal(t, "output", body)
}

func TestHTTPAdapter_Error(t *testing.T) {
	handler := HTTPAdapter(func(ctx context.Context, data []byte) ([]byte, error) {
		return nil, context.DeadlineExceeded
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader("input"))
	rw := httptest.NewRecorder()

	handler(rw, req)

	resp := rw.Result()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
