package core

import (
	"context"
	"net/http"
)

// Request describes one outbound backend operation.
type Request struct {
	Method string
	Path   string

	Headers     map[string]string
	Body        any
	BodyBytes   []byte
	ContentType string
	Accept      string

	ExpectNoResponse bool
}

// Response is the raw transport result returned by [Backend].
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Backend is the HTTP transport interface used by all SDK services.
// It handles authentication, retry, error parsing, and serialization.
//
// Implementations must be safe for concurrent use. The default implementation
// is [BackendImpl], constructed via [NewBackendImpl].
type Backend interface {
	// Do sends an HTTP request and returns the raw transport response.
	// JSON decoding and higher-level response handling happen in shared helpers
	// above this layer.
	Do(ctx context.Context, req Request) (*Response, error)
}
