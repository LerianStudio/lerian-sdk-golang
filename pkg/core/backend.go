package core

import "context"

// Backend is the HTTP transport interface used by all SDK services.
// It handles authentication, retry, error parsing, and serialization.
//
// Implementations must be safe for concurrent use. The default implementation
// is [BackendImpl], constructed via [NewBackendImpl].
type Backend interface {
	// Call sends an HTTP request and deserializes the JSON response into result.
	// If result is nil (e.g. for DELETE operations), the response body is
	// discarded after status-code validation.
	Call(ctx context.Context, method, path string, body, result any) error

	// CallWithHeaders is like Call but merges additional headers into the request.
	// These headers take precedence over default headers and are set after
	// authentication enrichment.
	CallWithHeaders(ctx context.Context, method, path string,
		headers map[string]string, body, result any) error

	// CallRaw sends an HTTP request and returns the raw response bytes
	// without attempting JSON deserialization. This is useful for endpoints
	// that return non-JSON content (CSV exports, binary data, etc.).
	CallRaw(ctx context.Context, method, path string, body any) ([]byte, error)
}
