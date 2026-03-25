package core

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/auth"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/observability"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/performance"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/retry"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// BackendConfig holds the configuration for constructing a [BackendImpl].
// All fields are optional except BaseURL; sensible defaults are applied
// for any field left at its zero value.
type BackendConfig struct {
	// BaseURL is the root URL of the API (e.g. "https://api.lerian.io/v1").
	// Paths supplied to requests are appended to this URL.
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

	// DefaultHeaders are merged into every outbound request. Per-call request
	// headers take precedence over these defaults.
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
// retry with exponential backoff, tenant-ID and idempotency-key injection,
// and structured error classification.
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

	b.httpClient = secureSDKHTTPClient(b.httpClient)

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

// Do sends an HTTP request and returns the raw transport response.
func (b *BackendImpl) Do(ctx context.Context, req Request) (*Response, error) {
	return b.doRequest(ctx, req)
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ Backend = (*BackendImpl)(nil)
