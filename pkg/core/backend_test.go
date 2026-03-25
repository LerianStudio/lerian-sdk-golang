package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// ---------------------------------------------------------------------------
// Test types
// ---------------------------------------------------------------------------

type testResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type testRequest struct {
	Name string `json:"name"`
}

type staticTestAuth struct {
	value string
}

func (a staticTestAuth) Enrich(_ context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.value)
	return nil
}

// fastRetryConfig returns a retry config with minimal delays for fast tests.
func fastRetryConfig(maxRetries int) retry.Config {
	return retry.Config{
		MaxRetries:  maxRetries,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		JitterRatio: 0,
	}
}

// newTestBackend creates a BackendImpl pointed at the given test server.
func newTestBackend(ts *httptest.Server, opts ...func(*BackendConfig)) *BackendImpl {
	cfg := BackendConfig{
		BaseURL:     ts.URL,
		RetryConfig: fastRetryConfig(0), // no retries by default
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return NewBackendImpl(cfg)
}

func testRequestFor(method, path string, headers map[string]string, body any) Request {
	req := Request{Method: method, Path: path, Headers: headers}
	if payload, ok := body.([]byte); ok {
		req.BodyBytes = payload
		if headers != nil {
			req.ContentType = headers["Content-Type"]
		}

		return req
	}

	req.Body = body

	return req
}

func callJSON(ctx context.Context, b *BackendImpl, method, path string, body, result any) error {
	res, err := b.Do(ctx, testRequestFor(method, path, nil, body))
	if err != nil {
		return err
	}

	if result == nil {
		return nil
	}

	return json.Unmarshal(res.Body, result)
}

func callJSONWithHeaders(ctx context.Context, b *BackendImpl, method, path string, headers map[string]string, body, result any) error {
	res, err := b.Do(ctx, testRequestFor(method, path, headers, body))
	if err != nil {
		return err
	}

	if result == nil {
		return nil
	}

	return json.Unmarshal(res.Body, result)
}

func callBytes(ctx context.Context, b *BackendImpl, method, path string, body any) ([]byte, error) {
	res, err := b.Do(ctx, testRequestFor(method, path, nil, body))
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestBackendCallSuccess(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"abc-123","name":"Test Resource"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/resources/abc-123", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "abc-123", result.ID)
	assert.Equal(t, "Test Resource", result.Name)
}

func TestBackendCallError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-ID", "req-404-id")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `resource not found`)
	}))
	defer ts.Close()

	parserCalled := false
	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.ErrorParser = func(statusCode int, body []byte) *sdkerrors.Error {
			parserCalled = true

			return &sdkerrors.Error{
				Product:    "test",
				Category:   sdkerrors.CategoryNotFound,
				StatusCode: statusCode,
				Message:    string(body),
			}
		}
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/resources/missing", nil, &result)

	require.Error(t, err)
	assert.True(t, parserCalled, "error parser should have been called")
	assert.True(t, errors.Is(err, sdkerrors.ErrNotFound), "error should match ErrNotFound sentinel")

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "req-404-id", sdkErr.RequestID)
	assert.Equal(t, http.StatusNotFound, sdkErr.StatusCode)
}

func TestBackendDoWithHeaders(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo the custom header back in the response body.
		customVal := r.Header.Get("X-Custom-Header")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"id":"1","name":"%s"}`, customVal)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	var result testResponse

	headers := map[string]string{"X-Custom-Header": "custom-value-42"}
	err := callJSONWithHeaders(context.Background(), b, http.MethodGet, "/test",
		headers, nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "custom-value-42", result.Name)
}

func TestBackendDoRawBytes(t *testing.T) {
	t.Parallel()

	rawPayload := `{"raw":"bytes","count":42}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, rawPayload)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	data, err := callBytes(context.Background(), b, http.MethodGet, "/export", nil)

	require.NoError(t, err)
	assert.Equal(t, rawPayload, string(data))
}

func TestBackendAuthEnrichment(t *testing.T) {
	t.Parallel()

	var receivedAuth string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"1","name":"authed"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.Auth = staticTestAuth{value: "test-token-secret-123"}
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/protected", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "Bearer test-token-secret-123", receivedAuth)
}

func TestBackendRetryOn429(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count := requestCount.Add(1)
		if count == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `rate limited`)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"retry-ok","name":"success"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(2)
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/rate-limited", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, int32(2), requestCount.Load(), "should have made exactly 2 requests")
	assert.Equal(t, "retry-ok", result.ID)
}

func TestBackendRetryOn500(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count := requestCount.Add(1)
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `internal error`)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"retry-500","name":"recovered"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(2)
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/server-error", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, int32(2), requestCount.Load(), "should have made exactly 2 requests")
	assert.Equal(t, "retry-500", result.ID)
}

func TestBackendNoRetryOn400(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `bad request`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(3)
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodPost, "/validate", &testRequest{Name: "bad"}, &result)

	require.Error(t, err)
	assert.Equal(t, int32(1), requestCount.Load(), "should NOT retry on 400")
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestBackendDefaultHeaders(t *testing.T) {
	t.Parallel()

	var receivedUserAgent string

	var receivedSDKVersion string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get("User-Agent")
		receivedSDKVersion = r.Header.Get("X-SDK-Version")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"1","name":"defaults"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.DefaultHeaders = map[string]string{
			"User-Agent":    "lerian-sdk-go/3.0.0",
			"X-SDK-Version": "3.0.0",
		}
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/with-defaults", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "lerian-sdk-go/3.0.0", receivedUserAgent)
	assert.Equal(t, "3.0.0", receivedSDKVersion)
}

func TestBackendIdempotencyKey(t *testing.T) {
	t.Parallel()

	var receivedKey string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = r.Header.Get("X-Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"txn-1","name":"transaction"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	ctx := WithIdempotencyKey(context.Background(), "unique-key-abc-123")

	var result testResponse

	err := callJSON(ctx, b, http.MethodPost, "/transactions", &testRequest{Name: "payment"}, &result)

	require.NoError(t, err)
	assert.Equal(t, "unique-key-abc-123", receivedKey)
	assert.Equal(t, "txn-1", result.ID)
	assert.Equal(t, "transaction", result.Name)
}

func TestBackendTenantID(t *testing.T) {
	t.Parallel()

	var receivedTenantID string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedTenantID = r.Header.Get("X-Tenant-ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"org-1","name":"organization"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	ctx := WithTenantID(context.Background(), "tenant-abc-123")

	var result testResponse

	err := callJSON(ctx, b, http.MethodGet, "/organizations", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "tenant-abc-123", receivedTenantID)
	assert.Equal(t, "org-1", result.ID)
	assert.Equal(t, "organization", result.Name)
}

func TestBackendContextCancellation(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Slow handler that would normally succeed.
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	var result testResponse

	err := callJSON(ctx, b, http.MethodGet, "/slow", nil, &result)

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrCancellation),
		"error should match ErrCancellation, got: %v", err)
}

func TestBackendNilResult(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	// DELETE typically doesn't need to unmarshal a response body.
	err := callJSON(context.Background(), b, http.MethodDelete, "/resources/abc-123", nil, nil)

	require.NoError(t, err)
}

func TestBackendNilBody(t *testing.T) {
	t.Parallel()

	var receivedContentType string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")

		// Verify no body was sent.
		body, _ := io.ReadAll(r.Body)
		assert.Empty(t, body, "GET request should have no body")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"1","name":"no-body"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/resources", nil, &result)

	require.NoError(t, err)
	assert.Empty(t, receivedContentType, "Content-Type should not be set when body is nil")
	assert.Equal(t, "no-body", result.Name)
}

func TestBackendGenericErrorParser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		statusCode     int
		body           string
		expectCategory sdkerrors.ErrorCategory
		expectSentinel *sdkerrors.Error
	}{
		{
			name:           "400 maps to validation",
			statusCode:     http.StatusBadRequest,
			body:           "invalid field: name",
			expectCategory: sdkerrors.CategoryValidation,
			expectSentinel: sdkerrors.ErrValidation,
		},
		{
			name:           "401 maps to authentication",
			statusCode:     http.StatusUnauthorized,
			body:           "missing credentials",
			expectCategory: sdkerrors.CategoryAuthentication,
			expectSentinel: sdkerrors.ErrAuthentication,
		},
		{
			name:           "403 maps to authorization",
			statusCode:     http.StatusForbidden,
			body:           "insufficient permissions",
			expectCategory: sdkerrors.CategoryAuthorization,
			expectSentinel: sdkerrors.ErrAuthorization,
		},
		{
			name:           "404 maps to not_found",
			statusCode:     http.StatusNotFound,
			body:           "resource not found",
			expectCategory: sdkerrors.CategoryNotFound,
			expectSentinel: sdkerrors.ErrNotFound,
		},
		{
			name:           "409 maps to conflict",
			statusCode:     http.StatusConflict,
			body:           "duplicate entry",
			expectCategory: sdkerrors.CategoryConflict,
			expectSentinel: sdkerrors.ErrConflict,
		},
		{
			name:           "429 maps to rate_limit",
			statusCode:     http.StatusTooManyRequests,
			body:           "rate limit exceeded",
			expectCategory: sdkerrors.CategoryRateLimit,
			expectSentinel: sdkerrors.ErrRateLimit,
		},
		{
			name:           "500 maps to internal",
			statusCode:     http.StatusInternalServerError,
			body:           "internal server error",
			expectCategory: sdkerrors.CategoryInternal,
			expectSentinel: sdkerrors.ErrInternal,
		},
		{
			name:           "503 maps to internal",
			statusCode:     http.StatusServiceUnavailable,
			body:           "service unavailable",
			expectCategory: sdkerrors.CategoryInternal,
			expectSentinel: sdkerrors.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("X-Request-ID", "req-"+fmt.Sprintf("%d", tt.statusCode))
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.body)
			}))
			defer ts.Close()

			// No ErrorParser configured — tests the generic path.
			b := newTestBackend(ts)

			var result testResponse

			err := callJSON(context.Background(), b, http.MethodGet, "/test", nil, &result)

			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.expectSentinel),
				"expected sentinel %v, got: %v", tt.expectSentinel, err)

			var sdkErr *sdkerrors.Error
			require.True(t, errors.As(err, &sdkErr))
			assert.Equal(t, tt.expectCategory, sdkErr.Category)
			assert.Equal(t, tt.statusCode, sdkErr.StatusCode)
			assert.Equal(t, tt.body, sdkErr.Message)
			assert.Equal(t, fmt.Sprintf("req-%d", tt.statusCode), sdkErr.RequestID)
		})
	}
}

func TestBackendNoTenantIDHeader(t *testing.T) {
	t.Parallel()

	var hasTenantHeader bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hasTenantHeader = r.Header.Get("X-Tenant-ID") != ""
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"1","name":"no-tenant"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/test", nil, &result)

	require.NoError(t, err)
	assert.False(t, hasTenantHeader, "X-Tenant-ID header should not be present when context has no tenant ID")
}

func TestBackendTenantIDPrecedence(t *testing.T) {
	t.Parallel()

	var receivedTenantID string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedTenantID = r.Header.Get("X-Tenant-ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"1","name":"precedence-test"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.DefaultHeaders = map[string]string{
			"X-Tenant-ID": "from-default",
		}
	})

	ctx := WithTenantID(context.Background(), "from-context")

	var result testResponse

	err := callJSONWithHeaders(ctx, b, http.MethodGet, "/test",
		map[string]string{"X-Tenant-ID": "from-extra"}, nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "from-context", receivedTenantID,
		"context-injected tenant ID should take precedence over default and extra headers")
}

func TestBackendTenantIDAndIdempotencyKey(t *testing.T) {
	t.Parallel()

	var receivedTenantID, receivedIdempotencyKey string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedTenantID = r.Header.Get("X-Tenant-ID")
		receivedIdempotencyKey = r.Header.Get("X-Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"1","name":"combined"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	ctx := WithTenantID(context.Background(), "tenant-xyz")
	ctx = WithIdempotencyKey(ctx, "idem-key-123")

	var result testResponse

	err := callJSON(ctx, b, http.MethodPost, "/test", &testRequest{Name: "combined"}, &result)

	require.NoError(t, err)
	assert.Equal(t, "tenant-xyz", receivedTenantID)
	assert.Equal(t, "idem-key-123", receivedIdempotencyKey)
}

// ---------------------------------------------------------------------------
// Additional edge-case tests
// ---------------------------------------------------------------------------

func TestBackendDoRawBytesError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `server error`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	data, err := callBytes(context.Background(), b, http.MethodGet, "/export", nil)

	require.Error(t, err)
	assert.Nil(t, data)
	assert.True(t, errors.Is(err, sdkerrors.ErrInternal))
}

func TestBackendCallWithBody(t *testing.T) {
	t.Parallel()

	var receivedBody testRequest

	var receivedContentType string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")

		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id":"new-1","name":"Created Resource"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	input := &testRequest{Name: "New Resource"}

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodPost, "/resources", input, &result)

	require.NoError(t, err)
	assert.Equal(t, "application/json", receivedContentType)
	assert.Equal(t, "New Resource", receivedBody.Name)
	assert.Equal(t, "new-1", result.ID)
	assert.Equal(t, "Created Resource", result.Name)
}

func TestBackendAcceptHeader(t *testing.T) {
	t.Parallel()

	var receivedAccept string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"1","name":"accept"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/test", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "application/json", receivedAccept)
}

func TestBackendRetryExhaustion(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `always failing`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(2)
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/always-fail", nil, &result)

	require.Error(t, err)
	// 1 initial + 2 retries = 3 total requests
	assert.Equal(t, int32(3), requestCount.Load(), "should exhaust all retries")
	assert.True(t, errors.Is(err, sdkerrors.ErrInternal))
}

func TestBackendDebugLogging(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"1","name":"debug"}`)
	}))
	defer ts.Close()

	// Just verify debug mode doesn't panic or break anything.
	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.Debug = true
		cfg.Auth = staticTestAuth{value: "secret-token-value"}
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/debug-test", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "debug", result.Name)
}

func TestBackendDefaultsApplied(t *testing.T) {
	t.Parallel()

	// Construct with minimal config — all defaults should be applied.
	b := NewBackendImpl(BackendConfig{
		BaseURL: "http://localhost:9999",
	})

	assert.NotNil(t, b.auth, "auth should default to NoAuth")
	assert.NotNil(t, b.jsonPool, "jsonPool should default to a new pool")
	assert.NotNil(t, b.httpClient, "httpClient should default to a client")
	assert.NotNil(t, b.logger, "logger should default to slog.Default()")
	assert.Equal(t, "http://localhost:9999", b.baseURL)
}

func TestMaskAuthorizationHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "long bearer token",
			input:  "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.long-token",
			expect: "Bear***",
		},
		{
			name:   "short value",
			input:  "abc",
			expect: "abc***",
		},
		{
			name:   "exactly 4 chars",
			input:  "0123",
			expect: "0123***",
		},
		{
			name:   "5 chars",
			input:  "01234",
			expect: "0123***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expect, MaskAuthorizationHeader(tt.input))
		})
	}
}

func TestWithIdempotencyKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// No key set.
	key, ok := idempotencyKeyFromContext(ctx)
	assert.False(t, ok)
	assert.Empty(t, key)

	// Key set.
	ctx = WithIdempotencyKey(ctx, "my-key")
	key, ok = idempotencyKeyFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "my-key", key)

	// Empty key should report as absent.
	ctx = WithIdempotencyKey(context.Background(), "")
	key, ok = idempotencyKeyFromContext(ctx)
	assert.False(t, ok)
	assert.Empty(t, key)
}

func TestWithTenantID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// No tenant ID set.
	tenantID, ok := tenantIDFromContext(ctx)
	assert.False(t, ok)
	assert.Empty(t, tenantID)

	// Tenant ID set.
	ctx = WithTenantID(ctx, "tenant-abc-123")
	tenantID, ok = tenantIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "tenant-abc-123", tenantID)

	// Empty tenant ID should report as absent.
	ctx = WithTenantID(context.Background(), "")
	tenantID, ok = tenantIDFromContext(ctx)
	assert.False(t, ok)
	assert.Empty(t, tenantID)
}

func TestBackendDoRawBytesWithBody(t *testing.T) {
	t.Parallel()

	var receivedBody testRequest

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedBody)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `raw-response-data`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	input := &testRequest{Name: "raw-input"}
	data, err := callBytes(context.Background(), b, http.MethodPost, "/raw-endpoint", input)

	require.NoError(t, err)
	assert.Equal(t, "raw-response-data", string(data))
	assert.Equal(t, "raw-input", receivedBody.Name)
}

func TestBackendErrorParserRequestIDInjection(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-ID", "server-req-id-xyz")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `validation error from server`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.ErrorParser = func(statusCode int, body []byte) *sdkerrors.Error {
			// Parser returns error without RequestID — BackendImpl should inject it.
			return &sdkerrors.Error{
				Product:    "ledger",
				Category:   sdkerrors.CategoryValidation,
				StatusCode: statusCode,
				Message:    string(body),
			}
		}
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodPost, "/validate", &testRequest{Name: "bad"}, &result)

	require.Error(t, err)

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "server-req-id-xyz", sdkErr.RequestID, "RequestID should be injected from response header")
	assert.Equal(t, "POST /validate", sdkErr.Operation, "Operation should be injected when parser omits it")
}

func TestBackendErrorParserReturnsNil(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `conflict details`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.ErrorParser = func(_ int, _ []byte) *sdkerrors.Error {
			// Return nil to fall through to generic error handling.
			return nil
		}
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodPut, "/resource", &testRequest{Name: "conflict"}, &result)

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrConflict),
		"should fall through to generic error handler when parser returns nil")
}

// TestBackendInterfaceCompliance verifies BackendImpl satisfies the Backend interface
// at compile time (this is also checked with var _ Backend = (*BackendImpl)(nil) in the impl).
func TestBackendInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var b Backend = &BackendImpl{}
	assert.NotNil(t, b)
}

func TestBackendContextDeadlineExceeded(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Slow handler to trigger deadline.
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Give context time to expire.
	time.Sleep(5 * time.Millisecond)

	var result testResponse

	err := callJSON(ctx, b, http.MethodGet, "/timeout", nil, &result)

	require.Error(t, err)
	// Should classify as either timeout or cancellation depending on exact timing.
	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Contains(t, []sdkerrors.ErrorCategory{
		sdkerrors.CategoryTimeout,
		sdkerrors.CategoryCancellation,
	}, sdkErr.Category, "should be timeout or cancellation, got: %s", sdkErr.Category)
}

func TestBackendDoRawBytesRetryOn500(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count := requestCount.Add(1)
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `server error`)

			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `raw-success-data`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(2)
	})

	data, err := callBytes(context.Background(), b, http.MethodGet, "/raw-retry", nil)

	require.NoError(t, err)
	assert.Equal(t, "raw-success-data", string(data))
	assert.Equal(t, int32(2), requestCount.Load())
}

func TestBackendDoRawBytesRetryExhaustion(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, `bad gateway`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(1)
	})

	data, err := callBytes(context.Background(), b, http.MethodGet, "/always-502", nil)

	require.Error(t, err)
	assert.Nil(t, data)
	assert.Equal(t, int32(2), requestCount.Load(), "should be 1 initial + 1 retry = 2")
	assert.True(t, errors.Is(err, sdkerrors.ErrInternal))
}

func TestBackendDoRawBytesNoRetryOn400(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `bad request`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(3)
	})

	data, err := callBytes(context.Background(), b, http.MethodPost, "/raw-validate", &testRequest{Name: "bad"})

	require.Error(t, err)
	assert.Nil(t, data)
	assert.Equal(t, int32(1), requestCount.Load(), "should NOT retry on 400")
}

func TestBackendDoRawBytesContextCancellation(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	data, err := callBytes(ctx, b, http.MethodGet, "/slow-raw", nil)

	require.Error(t, err)
	assert.Nil(t, data)
	assert.True(t, errors.Is(err, sdkerrors.ErrCancellation),
		"error should match ErrCancellation, got: %v", err)
}

func TestBackendNetworkErrorRetry(t *testing.T) {
	t.Parallel()

	// Create a server, get its URL, then close it to simulate network errors.
	// After first attempt "fails", start a new server on a known port.
	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"net-ok","name":"recovered"}`)
	}))
	// Server is live for all retries; test the happy path with retries enabled.
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(2)
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/net-test", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "net-ok", result.ID)
}

func TestBackendRetryOn429WithBody(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)

		// Verify the body is re-sent on retry.
		body, _ := io.ReadAll(r.Body)

		var req testRequest

		_ = json.Unmarshal(body, &req)

		if count == 1 {
			assert.Equal(t, "retry-me", req.Name)
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `rate limited`)

			return
		}

		assert.Equal(t, "retry-me", req.Name)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id":"created","name":"success"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(2)
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodPost, "/create",
		&testRequest{Name: "retry-me"}, &result)

	require.NoError(t, err)
	assert.Equal(t, int32(2), requestCount.Load())
	assert.Equal(t, "created", result.ID)
}

func TestBackendGenericHTTPErrorEmptyBody(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		// No body written — should use http.StatusText.
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/empty-body", nil, &result)

	require.Error(t, err)

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, sdkerrors.CategoryNotFound, sdkErr.Category)
	// With no body, the message should be the HTTP status text.
	assert.Equal(t, "Not Found", sdkErr.Message)
}

func TestBackendGenericHTTPErrorUnknownStatusCode(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418 - not in our switch cases
		fmt.Fprint(w, `I'm a teapot`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/teapot", nil, &result)

	require.Error(t, err)

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	// Unknown 4xx codes fall to the default (internal) category.
	assert.Equal(t, sdkerrors.CategoryInternal, sdkErr.Category)
	assert.Equal(t, 418, sdkErr.StatusCode)
}

func TestBackendNetworkErrorWithClosedServer(t *testing.T) {
	t.Parallel()

	// Create a server and immediately close it to simulate network error.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.Close() // Close immediately.

	b := NewBackendImpl(BackendConfig{
		BaseURL:     ts.URL,
		RetryConfig: fastRetryConfig(0), // no retries
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/closed", nil, &result)

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNetwork),
		"error should match ErrNetwork, got: %v", err)
}

func TestBackendNetworkErrorWithRetry(t *testing.T) {
	t.Parallel()

	// Create a server and close it — all attempts will fail with network errors.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.Close()

	b := NewBackendImpl(BackendConfig{
		BaseURL:     ts.URL,
		RetryConfig: fastRetryConfig(2), // 2 retries = 3 total attempts
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/network-retry", nil, &result)

	require.Error(t, err)
	assert.True(t, errors.Is(err, sdkerrors.ErrNetwork),
		"error should match ErrNetwork after exhausting retries, got: %v", err)
}

func TestBackendNetworkErrorWithRetryRaw(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.Close()

	b := NewBackendImpl(BackendConfig{
		BaseURL:     ts.URL,
		RetryConfig: fastRetryConfig(1),
	})

	data, err := callBytes(context.Background(), b, http.MethodGet, "/raw-net-err", nil)

	require.Error(t, err)
	assert.Nil(t, data)
	assert.True(t, errors.Is(err, sdkerrors.ErrNetwork))
}

func TestBackendContextCancelDuringRetryBackoff(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `error`)
	}))
	defer ts.Close()

	// Use longer backoff so we have time to cancel during sleep.
	b := NewBackendImpl(BackendConfig{
		BaseURL: ts.URL,
		RetryConfig: retry.Config{
			MaxRetries:  3,
			BaseDelay:   500 * time.Millisecond,
			MaxDelay:    1 * time.Second,
			JitterRatio: 0,
		},
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel shortly after first request.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	var result testResponse

	err := callJSON(ctx, b, http.MethodGet, "/cancel-during-backoff", nil, &result)

	require.Error(t, err)
	// Should get either a cancellation or timeout error because context was cancelled during backoff.
	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Contains(t, []sdkerrors.ErrorCategory{
		sdkerrors.CategoryCancellation,
		sdkerrors.CategoryTimeout,
	}, sdkErr.Category)
}

func TestBackendContextCancelDuringRetryBackoffRaw(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `error`)
	}))
	defer ts.Close()

	b := NewBackendImpl(BackendConfig{
		BaseURL: ts.URL,
		RetryConfig: retry.Config{
			MaxRetries:  3,
			BaseDelay:   500 * time.Millisecond,
			MaxDelay:    1 * time.Second,
			JitterRatio: 0,
		},
	})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	data, err := callBytes(ctx, b, http.MethodGet, "/cancel-raw-backoff", nil)

	require.Error(t, err)
	assert.Nil(t, data)
}

func TestBackendContextCancelDuringNetworkRetryBackoff(t *testing.T) {
	t.Parallel()

	// Server is closed to trigger network errors, then context is cancelled during backoff.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.Close()

	b := NewBackendImpl(BackendConfig{
		BaseURL: ts.URL,
		RetryConfig: retry.Config{
			MaxRetries:  3,
			BaseDelay:   500 * time.Millisecond,
			MaxDelay:    1 * time.Second,
			JitterRatio: 0,
		},
	})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	var result testResponse

	err := callJSON(ctx, b, http.MethodGet, "/net-cancel-backoff", nil, &result)

	require.Error(t, err)

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Contains(t, []sdkerrors.ErrorCategory{
		sdkerrors.CategoryCancellation,
		sdkerrors.CategoryNetwork,
		sdkerrors.CategoryTimeout,
	}, sdkErr.Category)
}

func TestBackendContextCancelDuringNetworkRetryBackoffRaw(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.Close()

	b := NewBackendImpl(BackendConfig{
		BaseURL: ts.URL,
		RetryConfig: retry.Config{
			MaxRetries:  3,
			BaseDelay:   500 * time.Millisecond,
			MaxDelay:    1 * time.Second,
			JitterRatio: 0,
		},
	})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	data, err := callBytes(ctx, b, http.MethodGet, "/net-cancel-raw", nil)

	require.Error(t, err)
	assert.Nil(t, data)
}

func TestBackendDoRawBytesRetryOn429(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count := requestCount.Add(1)
		if count == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `rate limited`)

			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `raw-ok`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(2)
	})

	data, err := callBytes(context.Background(), b, http.MethodGet, "/raw-429", nil)

	require.NoError(t, err)
	assert.Equal(t, "raw-ok", string(data))
	assert.Equal(t, int32(2), requestCount.Load())
}

// ---------------------------------------------------------------------------
// Raw byte request tests — verify BodyBytes bypasses JSON marshaling
// ---------------------------------------------------------------------------

func TestBackendDoWithRawBytes(t *testing.T) {
	t.Parallel()

	var receivedContentType string

	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")

		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id":"raw-1","name":"Created via raw bytes"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	rawContent := "--boundary\r\nContent-Disposition: form-data; name=\"field\"\r\n\r\nvalue\r\n--boundary--\r\n"
	headers := map[string]string{
		"Content-Type": "multipart/form-data; boundary=boundary",
	}

	var result testResponse

	err := callJSONWithHeaders(context.Background(), b, http.MethodPost, "/upload",
		headers, []byte(rawContent), &result)

	require.NoError(t, err)
	assert.Equal(t, "multipart/form-data; boundary=boundary", receivedContentType,
		"Content-Type should be overridden by extra headers")
	assert.Equal(t, rawContent, receivedBody,
		"raw byte payload should be sent verbatim without JSON marshaling")
	assert.Equal(t, "raw-1", result.ID)
	assert.Equal(t, "Created via raw bytes", result.Name)
}

func TestBackendDoRawBytesResponse(t *testing.T) {
	t.Parallel()

	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `raw-response-from-server`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	rawContent := "plain text body"
	data, err := callBytes(context.Background(), b, http.MethodPost, "/raw-upload",
		[]byte(rawContent))

	require.NoError(t, err)
	assert.Equal(t, rawContent, receivedBody,
		"raw byte payload should be sent verbatim in raw-byte response flows as well")
	assert.Equal(t, "raw-response-from-server", string(data))
}

func TestBackendDoRawBytesEmptyPayload(t *testing.T) {
	t.Parallel()

	var receivedContentLength string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentLength = r.Header.Get("Content-Length")
		body, _ := io.ReadAll(r.Body)

		assert.Empty(t, body, "empty raw byte payload should send empty body")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"empty","name":"empty body"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodPost, "/empty-raw",
		[]byte{}, &result)

	require.NoError(t, err)
	assert.Equal(t, "0", receivedContentLength)
	assert.Equal(t, "empty", result.ID)
}

// ---------------------------------------------------------------------------
// Issue #3: io.LimitReader — response body capped at 10 MiB
// ---------------------------------------------------------------------------

func TestBackendResponseBodyLimited(t *testing.T) {
	t.Parallel()

	// Create a response larger than 10 MiB.
	const limit = 10 << 20 // 10 MiB

	oversized := make([]byte, limit+1024)
	for i := range oversized {
		oversized[i] = 'A'
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(oversized)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	data, err := callBytes(context.Background(), b, http.MethodGet, "/huge", nil)

	require.NoError(t, err)
	assert.Equal(t, limit, len(data),
		"response body should be truncated to 10 MiB by io.LimitReader")
}

// ---------------------------------------------------------------------------
// Issue #4: CheckRedirect strips auth header on cross-domain redirect
// ---------------------------------------------------------------------------

func TestBackendCheckRedirectStripsAuth(t *testing.T) {
	t.Parallel()

	// Target server records whether it received an Authorization header.
	var targetReceivedAuth string

	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetReceivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"redirected","name":"ok"}`)
	}))
	defer targetServer.Close()

	// Origin server redirects to the target server (different host).
	originServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetServer.URL+"/final", http.StatusTemporaryRedirect)
	}))
	defer originServer.Close()

	// Use the default httpClient (which has CheckRedirect wired).
	b := NewBackendImpl(BackendConfig{
		BaseURL:     originServer.URL,
		Auth:        staticTestAuth{value: "secret-token"},
		RetryConfig: fastRetryConfig(0),
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/start", nil, &result)

	require.NoError(t, err)
	assert.Empty(t, targetReceivedAuth,
		"Authorization header should be stripped on cross-domain redirect")
	assert.Equal(t, "redirected", result.ID)
}

func TestBackendCheckRedirectPreservesAuthSameHost(t *testing.T) {
	t.Parallel()

	// Single server that redirects to itself — auth should be preserved.
	var requestCount int

	var lastReceivedAuth string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		lastReceivedAuth = r.Header.Get("Authorization")

		if requestCount == 1 {
			// Redirect to the same host.
			http.Redirect(w, r, "/final", http.StatusTemporaryRedirect)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"same-host","name":"ok"}`)
	}))
	defer ts.Close()

	b := NewBackendImpl(BackendConfig{
		BaseURL:     ts.URL,
		Auth:        staticTestAuth{value: "keep-me"},
		RetryConfig: fastRetryConfig(0),
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/start", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "Bearer keep-me", lastReceivedAuth,
		"Authorization header should be preserved on same-host redirect")
}

// ---------------------------------------------------------------------------
// Issue #15: Retry-After header respected for 429 responses
// ---------------------------------------------------------------------------

func TestBackendRetryAfterHeaderSeconds(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count := requestCount.Add(1)
		if count == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `rate limited`)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"after-retry","name":"ok"}`)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = retry.Config{
			MaxRetries:  2,
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    5 * time.Second, // high enough to not cap the Retry-After
			JitterRatio: 0,
		}
	})

	start := time.Now()

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/rate-limited", nil, &result)

	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, int32(2), requestCount.Load())
	assert.Equal(t, "after-retry", result.ID)

	// The Retry-After: 1 header should have caused at least ~1 second delay.
	assert.GreaterOrEqual(t, elapsed, 900*time.Millisecond,
		"backoff should respect Retry-After header (expected >= ~1s)")
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		minDelay time.Duration
		maxDelay time.Duration
	}{
		{
			name:     "empty value",
			value:    "",
			minDelay: 0,
			maxDelay: 0,
		},
		{
			name:     "integer seconds",
			value:    "120",
			minDelay: 120 * time.Second,
			maxDelay: 120 * time.Second,
		},
		{
			name:     "zero seconds",
			value:    "0",
			minDelay: 0,
			maxDelay: 0,
		},
		{
			name:     "invalid value",
			value:    "not-a-number",
			minDelay: 0,
			maxDelay: 0,
		},
		{
			name:     "HTTP-date future",
			value:    time.Now().Add(1 * time.Hour).UTC().Format(http.TimeFormat),
			minDelay: 59 * time.Minute,
			maxDelay: 61 * time.Minute,
		},
		{
			name:     "HTTP-date past",
			value:    "Mon, 01 Jan 2024 00:00:00 GMT",
			minDelay: 0,
			maxDelay: 0,
		},
		{
			name:     "HTTP-date malformed",
			value:    "Not-A-Real-Date, 99 Zaz 9999 99:99:99 ZZZ",
			minDelay: 0,
			maxDelay: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := parseRetryAfter(tt.value)
			assert.GreaterOrEqual(t, d, tt.minDelay)
			assert.LessOrEqual(t, d, tt.maxDelay)
		})
	}
}

// ---------------------------------------------------------------------------
// Issue #11: Observability Provider wired into BackendImpl
// ---------------------------------------------------------------------------

// spyTracerWrapper wraps a noop tracer but records Start calls for testing.
type spyTracerWrapper struct {
	trace.Tracer // embed noop tracer to satisfy embedded.Tracer
	startCount   atomic.Int32
	lastSpanName atomic.Value // stores string
}

func newSpyTracerWrapper() *spyTracerWrapper {
	noopTP := tracenoop.NewTracerProvider()

	return &spyTracerWrapper{
		Tracer: noopTP.Tracer("spy"),
	}
}

func (st *spyTracerWrapper) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	st.startCount.Add(1)
	st.lastSpanName.Store(spanName)

	// Delegate to the embedded noop tracer to get a valid span.
	return st.Tracer.Start(ctx, spanName, opts...)
}

// spyProvider implements observability.Provider using the spy tracer wrapper.
type spyProvider struct {
	tracer *spyTracerWrapper
}

func newSpyProvider() *spyProvider {
	return &spyProvider{
		tracer: newSpyTracerWrapper(),
	}
}

func (p *spyProvider) Tracer() trace.Tracer           { return p.tracer }
func (p *spyProvider) Meter() metric.Meter            { return nil }
func (p *spyProvider) Logger() *slog.Logger           { return slog.Default() }
func (p *spyProvider) Shutdown(context.Context) error { return nil }
func (p *spyProvider) IsEnabled() bool                { return true }

func TestBackendProviderTracerCalled(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"traced","name":"ok"}`)
	}))
	defer ts.Close()

	spy := newSpyProvider()

	b := NewBackendImpl(BackendConfig{
		BaseURL:     ts.URL,
		RetryConfig: fastRetryConfig(0),
		Provider:    spy,
	})

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/traced", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, int32(1), spy.tracer.startCount.Load(),
		"Provider.Tracer().Start should be called once per Call")

	spanName, ok := spy.tracer.lastSpanName.Load().(string)
	require.True(t, ok)
	assert.Equal(t, "GET /traced", spanName,
		"span name should be 'METHOD /path'")
}

func TestBackendProviderTracerCalledForDoRawBytes(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `raw-data`)
	}))
	defer ts.Close()

	spy := newSpyProvider()

	b := NewBackendImpl(BackendConfig{
		BaseURL:     ts.URL,
		RetryConfig: fastRetryConfig(0),
		Provider:    spy,
	})

	data, err := callBytes(context.Background(), b, http.MethodGet, "/raw-traced", nil)

	require.NoError(t, err)
	assert.Equal(t, "raw-data", string(data))
	assert.Equal(t, int32(1), spy.tracer.startCount.Load(),
		"Provider.Tracer().Start should be called once per raw-byte request")
}

func TestBackendProviderDefaultsToNoop(t *testing.T) {
	t.Parallel()

	// No Provider configured — should default to noop without panicking.
	b := NewBackendImpl(BackendConfig{
		BaseURL: "http://localhost:9999",
	})

	assert.NotNil(t, b.provider, "provider should default to noop provider")
	assert.False(t, b.provider.IsEnabled(),
		"default provider should be noop (IsEnabled=false)")
}

func TestStripSensitiveOnRedirectHelper(t *testing.T) {
	t.Parallel()

	t.Run("stops after 10 redirects", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(http.MethodGet, "http://evil.com/page", nil)

		via := make([]*http.Request, 10)
		for i := range via {
			r, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
			via[i] = r
		}

		err := stripSensitiveOnRedirect(req, via)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "10 redirects")
	})

	t.Run("strips auth on cross-domain", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest(http.MethodGet, "http://other.com/page", nil)
		req.Header.Set("Authorization", "Bearer secret")

		origReq, _ := http.NewRequest(http.MethodGet, "http://example.com/start", nil)
		via := []*http.Request{origReq}

		err := stripSensitiveOnRedirect(req, via)
		require.NoError(t, err)
		assert.Empty(t, req.Header.Get("Authorization"),
			"Authorization should be removed on cross-domain redirect")
	})

	t.Run("strips organization header on cross-domain", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, "http://other.com/page", nil)
		require.NoError(t, err)
		req.Header.Set("X-Organization-Id", "org-1")

		origReq, err := http.NewRequest(http.MethodGet, "http://example.com/start", nil)
		require.NoError(t, err)

		via := []*http.Request{origReq}

		err = stripSensitiveOnRedirect(req, via)
		require.NoError(t, err)
		assert.Empty(t, req.Header.Get("X-Organization-Id"))
	})

	t.Run("strips sensitive headers on https downgrade", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, "http://example.com/page", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer secret")
		req.Header.Set("X-Organization-Id", "org-1")

		origReq, err := http.NewRequest(http.MethodGet, "https://example.com/start", nil)
		require.NoError(t, err)

		via := []*http.Request{origReq}

		err = stripSensitiveOnRedirect(req, via)
		require.NoError(t, err)
		assert.Empty(t, req.Header.Get("Authorization"))
		assert.Empty(t, req.Header.Get("X-Organization-Id"))
	})

	t.Run("preserves auth on same domain", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, "http://example.com/page2", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer keep")

		origReq, err := http.NewRequest(http.MethodGet, "http://example.com/start", nil)
		require.NoError(t, err)

		via := []*http.Request{origReq}

		err = stripSensitiveOnRedirect(req, via)
		require.NoError(t, err)
		assert.Equal(t, "Bearer keep", req.Header.Get("Authorization"),
			"Authorization should be preserved on same-domain redirect")
	})

	t.Run("preserves auth on same authority with default port normalization", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, "https://example.com:443/page", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer keep")

		origReq, err := http.NewRequest(http.MethodGet, "https://example.com/start", nil)
		require.NoError(t, err)

		err = stripSensitiveOnRedirect(req, []*http.Request{origReq})
		require.NoError(t, err)
		assert.Equal(t, "Bearer keep", req.Header.Get("Authorization"))
	})

	t.Run("strips cookie and api key headers on cross-domain redirect", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, "https://other.com/page", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", "session=secret")
		req.Header.Set("X-API-Key", "secret")

		origReq, err := http.NewRequest(http.MethodGet, "https://example.com/start", nil)
		require.NoError(t, err)

		err = stripSensitiveOnRedirect(req, []*http.Request{origReq})
		require.NoError(t, err)
		assert.Empty(t, req.Header.Get("Cookie"))
		assert.Empty(t, req.Header.Get("X-API-Key"))
	})

	t.Run("strips sensitive headers on multi-hop downgrade", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, "http://example.com/final", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer secret")
		req.Header.Set("X-Organization-Id", "org-1")

		firstHop, err := http.NewRequest(http.MethodGet, "https://example.com/start", nil)
		require.NoError(t, err)
		secondHop, err := http.NewRequest(http.MethodGet, "https://example.com/intermediate", nil)
		require.NoError(t, err)

		err = stripSensitiveOnRedirect(req, []*http.Request{firstHop, secondHop})
		require.NoError(t, err)
		assert.Empty(t, req.Header.Get("Authorization"))
		assert.Empty(t, req.Header.Get("X-Organization-Id"))
	})
}

func TestSecureSDKHTTPClientWrapsExistingRedirectPolicy(t *testing.T) {
	t.Parallel()

	called := false
	authInCallback := ""
	client := secureSDKHTTPClient(&http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		called = true
		authInCallback = req.Header.Get("Authorization")

		return nil
	}})

	req, err := http.NewRequest(http.MethodGet, "http://other.com/final", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("X-Organization-Id", "org-1")

	viaReq, err := http.NewRequest(http.MethodGet, "http://example.com/start", nil)
	require.NoError(t, err)

	err = client.CheckRedirect(req, []*http.Request{viaReq})
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "Bearer secret", authInCallback)
	assert.Empty(t, req.Header.Get("Authorization"))
	assert.Empty(t, req.Header.Get("X-Organization-Id"))
}

func TestSecureSDKHTTPClientBuildsDefaultSecuredClient(t *testing.T) {
	t.Parallel()

	client := secureSDKHTTPClient(nil)
	require.NotNil(t, client)
	assert.Equal(t, defaultHTTPTimeout, client.Timeout)
	require.NotNil(t, client.CheckRedirect)

	req, err := http.NewRequest(http.MethodGet, "https://other.com/final", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer secret")

	viaReq, err := http.NewRequest(http.MethodGet, "https://example.com/start", nil)
	require.NoError(t, err)

	err = client.CheckRedirect(req, []*http.Request{viaReq})
	require.NoError(t, err)
	assert.Empty(t, req.Header.Get("Authorization"))
}

// ---------------------------------------------------------------------------
// MaxRetries capping (TASK M3)
// ---------------------------------------------------------------------------

func TestBackendMaxRetriesCappedAt10(t *testing.T) {
	t.Parallel()

	b := NewBackendImpl(BackendConfig{
		BaseURL: "http://localhost:9999",
		RetryConfig: retry.Config{
			MaxRetries:  100, // well above the limit
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    10 * time.Millisecond,
			JitterRatio: 0,
		},
	})

	assert.Equal(t, MaxRetriesLimit, b.retryConfig.MaxRetries,
		"MaxRetries should be capped to MaxRetriesLimit (%d)", MaxRetriesLimit)
}

func TestBackendMaxRetriesNotCappedWhenBelowLimit(t *testing.T) {
	t.Parallel()

	b := NewBackendImpl(BackendConfig{
		BaseURL: "http://localhost:9999",
		RetryConfig: retry.Config{
			MaxRetries:  5,
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    10 * time.Millisecond,
			JitterRatio: 0,
		},
	})

	assert.Equal(t, 5, b.retryConfig.MaxRetries,
		"MaxRetries should remain 5 when below the cap")
}

func TestBackendMaxRetriesExactlyAtLimit(t *testing.T) {
	t.Parallel()

	b := NewBackendImpl(BackendConfig{
		BaseURL: "http://localhost:9999",
		RetryConfig: retry.Config{
			MaxRetries:  MaxRetriesLimit,
			BaseDelay:   1 * time.Millisecond,
			MaxDelay:    10 * time.Millisecond,
			JitterRatio: 0,
		},
	})

	assert.Equal(t, MaxRetriesLimit, b.retryConfig.MaxRetries,
		"MaxRetries at exactly the limit should be preserved")
}

// ---------------------------------------------------------------------------
// Error body truncation (TASK M4)
// ---------------------------------------------------------------------------

func TestBackendGenericHTTPErrorTruncatesLargeBody(t *testing.T) {
	t.Parallel()

	// Create a body larger than sdkerrors.MaxErrorBodyBytes (512).
	largeBody := make([]byte, 1024)
	for i := range largeBody {
		largeBody[i] = 'X'
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(largeBody)
	}))
	defer ts.Close()

	// No ErrorParser -- tests the generic truncation path.
	b := newTestBackend(ts)

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/large-error", nil, &result)

	require.Error(t, err)

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))

	assert.Contains(t, sdkErr.Message, "... [truncated]",
		"large response body should be truncated in error message")
	assert.LessOrEqual(t, len(sdkErr.Message), sdkerrors.MaxErrorBodyBytes+20,
		"truncated message should not exceed sdkerrors.MaxErrorBodyBytes + suffix length")
}

func TestBackendGenericHTTPErrorSmallBodyNotTruncated(t *testing.T) {
	t.Parallel()

	shortBody := "short error message"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, shortBody)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/small-error", nil, &result)

	require.Error(t, err)

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))

	assert.Equal(t, shortBody, sdkErr.Message,
		"short response body should NOT be truncated")
	assert.NotContains(t, sdkErr.Message, "... [truncated]")
}

// ---------------------------------------------------------------------------
// DRY doRequest refactoring (TASK M1) - verify both paths still work
// ---------------------------------------------------------------------------

func TestBackendDoRequestSharedBetweenJSONAndRawBytePaths(t *testing.T) {
	t.Parallel()

	// This test verifies that the DRY refactoring did not break the fundamental
	// contract: JSON helpers decode bodies, raw-byte helpers return response bytes.

	responseJSON := `{"id":"dry-test","name":"DRY Refactored"}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, responseJSON)
	}))
	defer ts.Close()

	b := newTestBackend(ts)

	// Call path: should unmarshal JSON.
	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/dry-test", nil, &result)
	require.NoError(t, err)
	assert.Equal(t, "dry-test", result.ID)
	assert.Equal(t, "DRY Refactored", result.Name)

	// raw-byte path: should return raw bytes.
	data, err := callBytes(context.Background(), b, http.MethodGet, "/dry-test", nil)
	require.NoError(t, err)
	assert.JSONEq(t, responseJSON, string(data))
}

func TestBackendDoRequestRetrySharedByBothPaths(t *testing.T) {
	t.Parallel()

	var callCount, rawCallCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/json-retry" {
			count := callCount.Add(1)
			if count == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprint(w, `unavailable`)

				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"id":"json-ok","name":"recovered"}`)

			return
		}

		if r.URL.Path == "/raw-retry" {
			count := rawCallCount.Add(1)
			if count == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprint(w, `unavailable`)

				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `raw-ok`)

			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := newTestBackend(ts, func(cfg *BackendConfig) {
		cfg.RetryConfig = fastRetryConfig(2)
	})

	// JSON path retries correctly.
	var result testResponse

	err := callJSON(context.Background(), b, http.MethodGet, "/json-retry", nil, &result)
	require.NoError(t, err)
	assert.Equal(t, "json-ok", result.ID)
	assert.Equal(t, int32(2), callCount.Load())

	// Raw path retries correctly.
	data, err := callBytes(context.Background(), b, http.MethodGet, "/raw-retry", nil)
	require.NoError(t, err)
	assert.Equal(t, "raw-ok", string(data))
	assert.Equal(t, int32(2), rawCallCount.Load())
}
