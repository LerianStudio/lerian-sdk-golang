package core

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func backendHead(ctx context.Context, backend *BackendImpl, path string) (http.Header, error) {
	res, err := backend.Do(ctx, Request{Method: http.MethodHead, Path: path})
	if err != nil {
		return nil, err
	}

	return res.Headers.Clone(), nil
}

func TestBackendImplDoHead(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodHead, r.Method)
		assert.Equal(t, "/resources", r.URL.Path)
		w.Header().Set("X-Total-Count", "7")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	backend := NewBackendImpl(BackendConfig{BaseURL: server.URL, HTTPClient: server.Client()})
	headers, err := backendHead(context.Background(), backend, "/resources")
	require.NoError(t, err)
	assert.Equal(t, []string{"7"}, headers["X-Total-Count"])
}

func TestBackendImplDoHeadHTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodHead, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		_, err := fmt.Fprint(w, `{"message":"boom"}`)
		require.NoError(t, err)
	}))
	defer server.Close()

	backend := NewBackendImpl(BackendConfig{BaseURL: server.URL, HTTPClient: server.Client()})
	headers, err := backendHead(context.Background(), backend, "/resources")
	require.Error(t, err)
	assert.Nil(t, headers)
	assert.Contains(t, err.Error(), "HEAD /resources")
}

func TestBackendImplDoHeadRetriesRetryableResponses(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodHead, r.Method)
		assert.Equal(t, "/resources", r.URL.Path)

		if requestCount.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-Total-Count", "7")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	backend := NewBackendImpl(BackendConfig{
		BaseURL:     server.URL,
		HTTPClient:  server.Client(),
		RetryConfig: fastRetryConfig(1),
	})

	headers, err := backendHead(context.Background(), backend, "/resources")
	require.NoError(t, err)
	assert.Equal(t, []string{"7"}, headers["X-Total-Count"])
	assert.Equal(t, int32(2), requestCount.Load())
}
