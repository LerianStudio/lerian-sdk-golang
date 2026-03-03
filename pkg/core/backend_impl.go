package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/auth"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/observability"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/performance"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/retry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// BackendConfig holds the configuration for constructing a [BackendImpl].
// All fields are optional except BaseURL; sensible defaults are applied
// for any field left at its zero value.
type BackendConfig struct {
	// BaseURL is the root URL of the API (e.g. "https://api.lerian.io/v1").
	// Paths supplied to Call/CallRaw are appended to this URL.
	BaseURL string

	// Auth is the authentication strategy injected into every request.
	// Defaults to [auth.NewNoAuth] if nil.
	Auth auth.Authenticator

	// ErrorParser converts an HTTP error response into a structured SDK error.
	// It receives the status code and the raw response body.
	// If nil, BackendImpl creates a generic error from the status code and body.
	ErrorParser func(statusCode int, body []byte) *sdkerrors.Error

	// RetryConfig governs retry behaviour for transient failures (429, 5xx).
	// The zero value disables retries; use [retry.DefaultConfig] for defaults.
	RetryConfig retry.Config

	// JSONPool is the pooled JSON encoder/decoder used for request/response
	// serialization. Defaults to a fresh [performance.NewJSONPool] if nil.
	JSONPool *performance.JSONPool

	// Debug enables verbose request/response logging via the configured Logger.
	// Authorization header values are masked in debug output.
	Debug bool

	// DefaultHeaders are merged into every outbound request. Per-call headers
	// from CallWithHeaders take precedence over these defaults.
	DefaultHeaders map[string]string

	// HTTPClient is the underlying transport. Defaults to a client with a
	// 30-second timeout if nil.
	HTTPClient *http.Client

	// Logger receives debug-level log messages when Debug is true.
	// Defaults to [slog.Default] if nil.
	Logger *slog.Logger

	// Provider is the observability provider used for distributed tracing,
	// metrics, and structured logging. Defaults to [observability.NewNoopProvider]
	// if nil.
	Provider observability.Provider
}

// ---------------------------------------------------------------------------
// BackendImpl
// ---------------------------------------------------------------------------

// BackendImpl is the default [Backend] implementation. It manages the full
// HTTP request lifecycle including JSON serialization, authentication,
// retry with exponential backoff, idempotency-key injection, and structured
// error classification.
//
// BackendImpl is safe for concurrent use.
type BackendImpl struct {
	baseURL        string
	auth           auth.Authenticator
	errorParser    func(statusCode int, body []byte) *sdkerrors.Error
	retryConfig    retry.Config
	jsonPool       *performance.JSONPool
	debug          bool
	defaultHeaders map[string]string
	httpClient     *http.Client
	logger         *slog.Logger
	provider       observability.Provider
}

// defaultHTTPTimeout is the timeout applied to the default HTTP client
// when the caller does not supply one.
const defaultHTTPTimeout = 30 * time.Second

// MaxRetriesLimit is the hard upper bound for retry attempts. Any
// MaxRetries value exceeding this limit is silently capped during
// backend construction to prevent accidental runaway retry loops.
const MaxRetriesLimit = 10

// NewBackendImpl creates a [BackendImpl] from the given configuration,
// applying sensible defaults for any unset fields.
func NewBackendImpl(cfg BackendConfig) *BackendImpl {
	b := &BackendImpl{
		baseURL:        cfg.BaseURL,
		auth:           cfg.Auth,
		errorParser:    cfg.ErrorParser,
		retryConfig:    cfg.RetryConfig,
		jsonPool:       cfg.JSONPool,
		debug:          cfg.Debug,
		defaultHeaders: cfg.DefaultHeaders,
		httpClient:     cfg.HTTPClient,
		logger:         cfg.Logger,
	}

	// Cap MaxRetries at the hard upper bound to prevent runaway retry loops.
	if b.retryConfig.MaxRetries > MaxRetriesLimit {
		b.retryConfig.MaxRetries = MaxRetriesLimit
	}

	if b.auth == nil {
		b.auth = auth.NewNoAuth()
	}

	if b.jsonPool == nil {
		b.jsonPool = performance.NewJSONPool()
	}

	if b.httpClient == nil {
		b.httpClient = &http.Client{
			Timeout:       defaultHTTPTimeout,
			CheckRedirect: stripAuthOnRedirect,
		}
	}

	if b.logger == nil {
		b.logger = slog.Default()
	}

	b.provider = cfg.Provider
	if b.provider == nil {
		b.provider = observability.NewNoopProvider()
	}

	return b
}

// ---------------------------------------------------------------------------
// Backend interface implementation
// ---------------------------------------------------------------------------

// Call sends an HTTP request and deserializes the JSON response into result.
func (b *BackendImpl) Call(ctx context.Context, method, path string, body, result any) error {
	return b.doCall(ctx, method, path, nil, body, result)
}

// CallWithHeaders is like Call but merges additional headers into the request.
func (b *BackendImpl) CallWithHeaders(ctx context.Context, method, path string,
	headers map[string]string, body, result any) error {
	return b.doCall(ctx, method, path, headers, body, result)
}

// CallRaw sends an HTTP request and returns the raw response bytes.
func (b *BackendImpl) CallRaw(ctx context.Context, method, path string, body any) ([]byte, error) {
	return b.doCallRaw(ctx, method, path, nil, body)
}

// ---------------------------------------------------------------------------
// Internal: doRequest (shared retry loop)
// ---------------------------------------------------------------------------

// requestResult holds the raw HTTP result from a single successful request
// exchange (status 2xx/3xx). Error responses (4xx/5xx) are handled inside
// doRequest and never surface here; they are returned as *sdkerrors.Error.
type requestResult struct {
	body       []byte
	statusCode int
}

// doRequest is the unified retry loop shared by doCall and doCallRaw.
// It handles:
//   - Body marshaling (JSON or RawBody pass-through)
//   - The full retry loop with OTel span creation, request building,
//     HTTP execution, Retry-After parsing, and exponential backoff
//   - Reading and capping the response body (10 MiB limit)
//
// On success (2xx/3xx) it returns a requestResult with the raw body.
// On error it returns a structured *sdkerrors.Error.
func (b *BackendImpl) doRequest(ctx context.Context, method, path string,
	extraHeaders map[string]string, body any) (*requestResult, error) {
	operation := method + " " + path

	// Start an observability span for the entire call lifecycle.
	ctx, span := b.provider.Tracer().Start(ctx, operation)
	defer span.End()

	// 1. Marshal body to JSON if present. If the body is a RawBody,
	// use its bytes directly without JSON encoding -- this supports
	// non-JSON content types such as multipart/form-data.
	var bodyBytes []byte

	if body != nil {
		if raw, ok := body.(RawBody); ok {
			bodyBytes = raw.Data
		} else {
			var err error

			bodyBytes, err = b.jsonPool.Marshal(body)
			if err != nil {
				return nil, sdkerrors.NewInternal("sdk", operation, "failed to marshal request body", err)
			}
		}
	}

	// 2. Execute with retry loop.
	var lastErr error

	for attempt := 0; attempt <= b.retryConfig.MaxRetries; attempt++ {
		// Rebuild the request on each attempt because the body reader is consumed.
		req, err := b.buildRequest(ctx, method, path, extraHeaders, bodyBytes)
		if err != nil {
			return nil, err
		}

		b.logRequest(method, path, req)

		// Send the request.
		resp, err := b.httpClient.Do(req)
		if err != nil {
			lastErr = b.classifyNetworkError(ctx, operation, err)

			// Network errors are only retryable if the context is not done.
			if ctx.Err() != nil {
				return nil, lastErr
			}

			if attempt < b.retryConfig.MaxRetries {
				if sleepErr := b.backoffSleep(ctx, attempt, 0); sleepErr != nil {
					return nil, b.classifyNetworkError(ctx, operation, sleepErr)
				}

				continue
			}

			return nil, lastErr
		}

		// Read the response body (capped at 10 MiB to prevent OOM on
		// pathological responses).
		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
		resp.Body.Close()

		if readErr != nil {
			span.SetStatus(codes.Error, "failed to read response body")

			return nil, sdkerrors.NewInternal("sdk", operation, "failed to read response body", readErr)
		}

		b.logResponse(resp.StatusCode, path)

		// Set span attributes now that we have a response.
		span.SetAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", b.baseURL+path),
			attribute.Int("http.status_code", resp.StatusCode),
		)

		if resp.StatusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		}

		// Handle error responses (4xx / 5xx).
		if resp.StatusCode >= 400 {
			requestID := resp.Header.Get("X-Request-ID")

			// Parse Retry-After header for 429 responses so the backoff
			// respects the server's requested delay.
			var retryAfter time.Duration
			if resp.StatusCode == http.StatusTooManyRequests {
				retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
			}

			// Check if retryable and retries remain.
			if retry.IsRetryable(resp.StatusCode) && attempt < b.retryConfig.MaxRetries {
				lastErr = b.buildHTTPError(resp.StatusCode, respBody, requestID, operation)

				if sleepErr := b.backoffSleep(ctx, attempt, retryAfter); sleepErr != nil {
					return nil, b.classifyNetworkError(ctx, operation, sleepErr)
				}

				continue
			}

			return nil, b.buildHTTPError(resp.StatusCode, respBody, requestID, operation)
		}

		// Success: return the raw response body and status code.
		return &requestResult{body: respBody, statusCode: resp.StatusCode}, nil
	}

	// All retries exhausted.
	if lastErr != nil {
		return nil, lastErr
	}

	return nil, sdkerrors.NewInternal("sdk", operation, "all retry attempts exhausted", nil)
}

// ---------------------------------------------------------------------------
// Internal: doCall (JSON response path)
// ---------------------------------------------------------------------------

// doCall sends the HTTP request via doRequest and unmarshals the JSON
// response into result. If result is nil (e.g. DELETE with no body),
// the response bytes are discarded.
func (b *BackendImpl) doCall(ctx context.Context, method, path string,
	extraHeaders map[string]string, body, result any) error {
	res, err := b.doRequest(ctx, method, path, extraHeaders, body)
	if err != nil {
		return err
	}

	// Success path: if result is nil (e.g. DELETE), just return.
	if result == nil {
		return nil
	}

	operation := method + " " + path

	// Unmarshal the response body into result.
	if err := b.jsonPool.Unmarshal(res.body, result); err != nil {
		return sdkerrors.NewInternal("sdk", operation, "failed to unmarshal response body", err)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Internal: doCallRaw (raw bytes response path)
// ---------------------------------------------------------------------------

// doCallRaw sends the HTTP request via doRequest and returns the raw
// response bytes without any JSON unmarshaling.
func (b *BackendImpl) doCallRaw(ctx context.Context, method, path string,
	extraHeaders map[string]string, body any) ([]byte, error) {
	res, err := b.doRequest(ctx, method, path, extraHeaders, body)
	if err != nil {
		return nil, err
	}

	return res.body, nil
}

// ---------------------------------------------------------------------------
// Request building
// ---------------------------------------------------------------------------

// buildRequest constructs an [http.Request] with all headers applied:
// Content-Type, Accept, default headers, extra headers, idempotency key,
// and authentication.
func (b *BackendImpl) buildRequest(ctx context.Context, method, path string,
	extraHeaders map[string]string, bodyBytes []byte) (*http.Request, error) {
	operation := method + " " + path
	url := b.baseURL + path

	var bodyReader io.Reader
	if bodyBytes != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, sdkerrors.NewInternal("sdk", operation, "failed to create HTTP request", err)
	}

	// Set standard headers.
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")

	// Apply default headers from config.
	for k, v := range b.defaultHeaders {
		req.Header.Set(k, v)
	}

	// Apply per-call extra headers (takes precedence over defaults).
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	// Inject idempotency key from context if present.
	if key, ok := idempotencyKeyFromContext(ctx); ok {
		req.Header.Set("X-Idempotency-Key", key)
	}

	// Authenticate the request.
	if err := b.auth.Enrich(ctx, req); err != nil {
		return nil, sdkerrors.NewAuthentication("sdk", operation,
			fmt.Sprintf("authentication enrichment failed: %v", err))
	}

	// Propagate W3C trace context so downstream services can
	// correlate spans within the same distributed trace.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	return req, nil
}

// ---------------------------------------------------------------------------
// Error classification
// ---------------------------------------------------------------------------

// classifyNetworkError inspects the context and underlying error to produce
// the appropriate structured SDK error.
func (b *BackendImpl) classifyNetworkError(ctx context.Context, operation string, err error) *sdkerrors.Error {
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return sdkerrors.NewCancellation("sdk", operation, ctx.Err())
		}

		return sdkerrors.NewTimeout("sdk", operation, ctx.Err())
	}

	return sdkerrors.NewNetwork("sdk", operation, err)
}

// buildHTTPError constructs a structured SDK error from an HTTP error response.
// If an ErrorParser is configured, it is used; otherwise a generic error is
// created based on the status code.
func (b *BackendImpl) buildHTTPError(statusCode int, body []byte, requestID, operation string) *sdkerrors.Error {
	if b.errorParser != nil {
		sdkErr := b.errorParser(statusCode, body)
		if sdkErr != nil {
			if sdkErr.RequestID == "" {
				sdkErr.RequestID = requestID
			}

			if sdkErr.Operation == "" {
				sdkErr.Operation = operation
			}

			return sdkErr
		}
	}

	// Generic error construction when no parser is configured or parser returned nil.
	return b.genericHTTPError(statusCode, body, requestID, operation)
}

// genericHTTPError creates a structured error from the HTTP status code when
// no ErrorParser is available. It maps common status codes to the appropriate
// error category. Response bodies exceeding sdkerrors.MaxErrorBodyBytes are truncated
// to prevent enormous payloads from polluting logs and error output.
func (b *BackendImpl) genericHTTPError(statusCode int, body []byte, requestID, operation string) *sdkerrors.Error {
	message := http.StatusText(statusCode)
	if len(body) > 0 {
		message = string(body)
		if len(message) > sdkerrors.MaxErrorBodyBytes {
			message = message[:sdkerrors.MaxErrorBodyBytes] + "... [truncated]"
		}
	}

	sdkErr := &sdkerrors.Error{
		Product:    "sdk",
		Operation:  operation,
		StatusCode: statusCode,
		RequestID:  requestID,
		Message:    message,
	}

	switch {
	case statusCode == http.StatusBadRequest:
		sdkErr.Category = sdkerrors.CategoryValidation
	case statusCode == http.StatusUnauthorized:
		sdkErr.Category = sdkerrors.CategoryAuthentication
	case statusCode == http.StatusForbidden:
		sdkErr.Category = sdkerrors.CategoryAuthorization
	case statusCode == http.StatusNotFound:
		sdkErr.Category = sdkerrors.CategoryNotFound
	case statusCode == http.StatusConflict:
		sdkErr.Category = sdkerrors.CategoryConflict
	case statusCode == http.StatusTooManyRequests:
		sdkErr.Category = sdkerrors.CategoryRateLimit
	case statusCode >= 500:
		sdkErr.Category = sdkerrors.CategoryInternal
	default:
		sdkErr.Category = sdkerrors.CategoryInternal
	}

	return sdkErr
}

// ---------------------------------------------------------------------------
// Redirect safety
// ---------------------------------------------------------------------------

// stripAuthOnRedirect removes the Authorization header when following
// redirects to a different host, preventing credential leakage.
func stripAuthOnRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("stopped after 10 redirects")
	}

	if len(via) > 0 && req.URL.Host != via[0].URL.Host {
		req.Header.Del("Authorization")
	}

	return nil
}

// ---------------------------------------------------------------------------
// Retry backoff
// ---------------------------------------------------------------------------

// backoffSleep computes the exponential backoff delay for the given attempt
// and sleeps, honouring context cancellation.
//
// When retryAfter is non-zero (e.g. parsed from a Retry-After header) the
// sleep duration is the maximum of the computed exponential delay and
// retryAfter, capped at MaxDelay.
//
// The core delay calculation is delegated to [retry.CalculateDelay] to
// maintain a single source of truth for the exponential-backoff-with-jitter
// algorithm. This method adds Retry-After awareness and context-aware sleep.
func (b *BackendImpl) backoffSleep(ctx context.Context, attempt int, retryAfter time.Duration) error {
	delay := retry.CalculateDelay(b.retryConfig, attempt)

	// If the server requested a longer delay via Retry-After, honour it.
	if retryAfter > delay {
		delay = retryAfter
	}

	// Cap at the configured maximum delay (retryAfter may exceed it).
	if delay > b.retryConfig.MaxDelay {
		delay = b.retryConfig.MaxDelay
	}

	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// parseRetryAfter parses the value of an HTTP Retry-After header.
// It first tries to interpret the value as an integer number of seconds
// (most common for APIs), then falls back to HTTP-date format.
// Returns 0 if the value is empty or cannot be parsed.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}

	// Try seconds first (most common for APIs).
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try HTTP-date format.
	if t, err := http.ParseTime(value); err == nil {
		delay := time.Until(t)
		if delay > 0 {
			return delay
		}
	}

	return 0
}

// ---------------------------------------------------------------------------
// Debug logging
// ---------------------------------------------------------------------------

// logRequest logs the outbound request details when debug mode is enabled.
// Authorization header values are masked to avoid leaking credentials.
func (b *BackendImpl) logRequest(method, path string, req *http.Request) {
	if !b.debug {
		return
	}

	attrs := []any{
		"method", method,
		"path", path,
	}

	if authVal := req.Header.Get("Authorization"); authVal != "" {
		attrs = append(attrs, "authorization", maskHeader(authVal))
	}

	b.logger.Debug("sdk request", attrs...)
}

// logResponse logs the response status when debug mode is enabled.
func (b *BackendImpl) logResponse(statusCode int, path string) {
	if !b.debug {
		return
	}

	b.logger.Debug("sdk response", "status", statusCode, "path", path)
}

// maskHeader returns the first 4 characters of the header value followed
// by "***" to prevent credential leakage in logs. The low threshold ensures
// that Bearer token material is never exposed (e.g. "Bear***"). If the value
// is shorter than 4 characters, the entire value is shown followed by "***".
func maskHeader(value string) string {
	const visibleChars = 4

	if len(value) <= visibleChars {
		return value + "***"
	}

	return value[:visibleChars] + "***"
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ Backend = (*BackendImpl)(nil)

// ---------------------------------------------------------------------------
// Exported helpers for testing and extension
// ---------------------------------------------------------------------------

// MaskAuthorizationHeader is an exported wrapper around the internal
// maskHeader function, useful for custom debug middleware.
func MaskAuthorizationHeader(value string) string {
	return maskHeader(value)
}
