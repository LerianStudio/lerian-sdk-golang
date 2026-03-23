package lerian

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/auth"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/observability"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/performance"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/retry"
	"github.com/LerianStudio/lerian-sdk-golang/reporter"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
)

// sdkVersion is the SDK version string attached to the OTel resource.
const sdkVersion = "0.1.0"

// defaultHTTPTimeout is the timeout applied to the default HTTP client
// when the caller does not supply one.
const defaultHTTPTimeout = 30 * time.Second

// Client is the top-level Lerian SDK client. Create one with [New].
// Access product APIs via the exported product fields (e.g., client.Midaz,
// client.Matcher). Products not configured via With<Product>() options will
// be nil.
type Client struct {
	// Product clients — nil unless configured via the corresponding
	// With<Product> option during New().
	Midaz    *midaz.Client
	Matcher  *matcher.Client
	Tracer   *tracer.Client
	Reporter *reporter.Client
	Fees     *fees.Client

	// Shared infrastructure
	observability observability.Provider
	httpClient    *http.Client
	retryConfig   retry.Config
	jsonPool      *performance.JSONPool
	debug         bool
	debugExplicit bool // true when WithDebug was called (distinguishes false from "not set")
	shutdownOnce  sync.Once

	// Deferred product configs — populated by options, consumed during New().
	// The "requested" flags track whether With<Product>() was called, even
	// with no options (env vars can supply the values).
	midazOpts         []midaz.Option
	midazRequested    bool
	matcherOpts       []matcher.Option
	matcherRequested  bool
	tracerOpts        []tracer.Option
	tracerRequested   bool
	reporterOpts      []reporter.Option
	reporterRequested bool
	feesOpts          []fees.Option
	feesRequested     bool

	// Observability config — populated by options, consumed during New().
	otelTraces   bool
	otelMetrics  bool
	otelLogs     bool
	otelEndpoint string
}

// New constructs a fully-wired [Client] by applying the supplied functional
// options in order. Each option may configure a product client, override
// transport defaults, or attach observability providers.
//
// The construction sequence:
//  1. Create a default Client with sensible defaults (retry, JSON pool, HTTP client).
//  2. Apply all options (populating internal fields and deferred product configs).
//  3. Initialize the observability provider if any OTel flags are enabled.
//  4. Initialize each configured product, creating Backend instances and validating config.
//  5. Return the wired client or the first error encountered.
//
// Products are only initialized when their corresponding With<Product>()
// option has been supplied. Unconfigured products remain nil.
func New(opts ...Option) (*Client, error) {
	// 1. Create default client.
	c := &Client{
		retryConfig: retry.DefaultConfig(),
		jsonPool:    performance.NewJSONPool(),
		httpClient:  &http.Client{Timeout: defaultHTTPTimeout},
	}

	// 2. Apply all options.
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("lerian: applying option: %w", err)
		}
	}

	// 2b. Apply environment variable fallback for debug mode.
	// Explicit WithDebug() takes precedence (even WithDebug(false));
	// env var is only read when WithDebug was never called.
	if !c.debugExplicit {
		c.debug = envBool(envDebug)
	}

	// 3. Initialize observability.
	if c.otelTraces || c.otelMetrics || c.otelLogs {
		provider, err := observability.NewProvider(observability.ProviderConfig{
			ServiceName:       "lerian-sdk-go",
			ServiceVersion:    sdkVersion,
			CollectorEndpoint: c.otelEndpoint,
			EnableTraces:      c.otelTraces,
			EnableMetrics:     c.otelMetrics,
			EnableLogs:        c.otelLogs,
		})
		if err != nil {
			return nil, fmt.Errorf("lerian: initializing observability: %w", err)
		}

		c.observability = provider
	} else {
		c.observability = observability.NewNoopProvider()
	}

	// 4. Initialize configured products.
	if err := c.initMidaz(); err != nil {
		return nil, err
	}

	if err := c.initMatcher(); err != nil {
		return nil, err
	}

	if err := c.initTracer(); err != nil {
		return nil, err
	}

	if err := c.initReporter(); err != nil {
		return nil, err
	}

	if err := c.initFees(); err != nil {
		return nil, err
	}

	return c, nil
}

// Shutdown gracefully drains buffered telemetry and releases resources.
// It is safe to call Shutdown multiple times; only the first call
// performs work.
func (c *Client) Shutdown(ctx context.Context) error {
	var err error

	c.shutdownOnce.Do(func() {
		if c.observability != nil {
			err = c.observability.Shutdown(ctx)
		}
	})

	return err
}

// ---------------------------------------------------------------------------
// Internal: createBackend helper
// ---------------------------------------------------------------------------

// createBackend constructs a [core.BackendImpl] from shared infrastructure
// settings plus per-product overrides (base URL, authenticator, default
// headers, error parser, and optional per-product timeout).
//
// When timeout > 0, a shallow clone of the shared HTTP client is created
// with the overridden Timeout value so other products are not affected.
func (c *Client) createBackend(
	baseURL string,
	authenticator auth.Authenticator,
	errorParser func(int, []byte) *sdkerrors.Error,
	defaultHeaders map[string]string,
	timeout time.Duration,
) *core.BackendImpl {
	httpClient := c.httpClientForTimeout(timeout)

	return core.NewBackendImpl(core.BackendConfig{
		BaseURL:        baseURL,
		Auth:           authenticator,
		ErrorParser:    errorParser,
		RetryConfig:    c.retryConfig,
		JSONPool:       c.jsonPool,
		Debug:          c.debug,
		DefaultHeaders: defaultHeaders,
		HTTPClient:     httpClient,
		Provider:       c.observability,
	})
}

func cloneHTTPClient(base *http.Client, timeout time.Duration) *http.Client {
	cloned := *base
	if timeout > 0 {
		cloned.Timeout = timeout
	}

	return &cloned
}

func (c *Client) httpClientForTimeout(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		return c.httpClient
	}

	return cloneHTTPClient(c.httpClient, timeout)
}

// ---------------------------------------------------------------------------
// Internal: insecure URL warning
// ---------------------------------------------------------------------------

// isLocalhostURL reports whether rawURL targets a loopback address
// (localhost, 127.0.0.1, or [::1]), regardless of port or path.
// It returns false (not localhost) if the URL cannot be parsed, so that
// the warning fires for malformed URLs as well.
func isLocalhostURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	host := u.Hostname() // strips port and brackets

	return strings.EqualFold(host, "localhost") ||
		host == "127.0.0.1" ||
		host == "::1"
}

// warnInsecureURL emits a structured warning via [slog.Default] when
// rawURL uses the HTTP scheme and does not target a loopback address.
// This helps developers notice accidental plain-text transport in
// production configurations without blocking them.
func warnInsecureURL(product, rawURL string) {
	if !strings.HasPrefix(strings.ToLower(rawURL), "http://") {
		return // HTTPS or other scheme — no warning needed
	}

	if isLocalhostURL(rawURL) {
		return // localhost is expected during development
	}

	slog.Warn("lerian: insecure URL detected for "+product+": "+
		"using HTTP instead of HTTPS is not recommended for production use",
		"product", product,
		"url", rawURL,
	)
}

func resolveEnvOAuthConfig(envClientID, envClientSecret, envTokenURL string) (clientID, clientSecret, tokenURL string) {
	clientID = envOrDefault(envClientID, "")
	clientSecret = envOrDefault(envClientSecret, "")
	tokenURL = envOrDefault(envTokenURL, "")

	return clientID, clientSecret, tokenURL
}

func applyOAuthEnvFallbacks(clientID, clientSecret, tokenURL *string,
	envClientID, envClientSecret, envTokenURL string) {
	if *clientID != "" || *clientSecret != "" || *tokenURL != "" {
		return
	}

	envResolvedClientID, envResolvedClientSecret, envResolvedTokenURL := resolveEnvOAuthConfig(
		envClientID,
		envClientSecret,
		envTokenURL,
	)
	*clientID = envResolvedClientID
	*clientSecret = envResolvedClientSecret
	*tokenURL = envResolvedTokenURL
}

func buildOAuthAuthenticator(clientID, clientSecret, tokenURL string, scopes []string, httpClient *http.Client) auth.Authenticator {
	if clientID != "" && clientSecret != "" && tokenURL != "" {
		return auth.NewOAuth2WithHTTPClient(clientID, clientSecret, tokenURL, scopes, httpClient)
	}

	return auth.NewNoAuth()
}

func validateOAuthTokenURL(product, tokenURL string) error {
	if tokenURL == "" || isLocalhostURL(tokenURL) {
		return nil
	}

	if strings.HasPrefix(strings.ToLower(tokenURL), "http://") {
		return fmt.Errorf("lerian: %s: TokenURL must use HTTPS outside localhost", product)
	}

	return nil
}

func validateLegacyAuthEnvUnset(product, migrationHint string, legacyKeys ...string) error {
	for _, legacyKey := range legacyKeys {
		if envOrDefault(legacyKey, "") == "" {
			continue
		}

		return fmt.Errorf("lerian: %s: %s is no longer supported; migrate to %s", product, legacyKey, migrationHint)
	}

	return nil
}

func validateOAuthAuthConfig(product, clientID, clientSecret, tokenURL string) error {
	hasOAuthConfig := clientID != "" || clientSecret != "" || tokenURL != ""
	hasCompleteOAuthConfig := clientID != "" && clientSecret != "" && tokenURL != ""

	if hasOAuthConfig && !hasCompleteOAuthConfig {
		return fmt.Errorf("lerian: %s: ClientID, ClientSecret, and TokenURL must all be set for OAuth2", product)
	}

	return validateOAuthTokenURL(product, tokenURL)
}

func applyMidazEnvFallbacks(cfg *midaz.Config) {
	if cfg.OnboardingURL == "" {
		cfg.OnboardingURL = envOrDefault(envMidazOnboardingURL, "")
	}

	if cfg.TransactionURL == "" {
		cfg.TransactionURL = envOrDefault(envMidazTransactionURL, "")
	}

	applyOAuthEnvFallbacks(
		&cfg.ClientID,
		&cfg.ClientSecret,
		&cfg.TokenURL,
		envMidazClientID,
		envMidazClientSecret,
		envMidazTokenURL,
	)
}

func buildMidazAuthenticator(cfg midaz.Config, httpClient *http.Client) auth.Authenticator {
	return buildOAuthAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.TokenURL, cfg.Scopes, httpClient)
}

func validateMidazAuthConfig(cfg midaz.Config) error {
	return validateOAuthAuthConfig("midaz", cfg.ClientID, cfg.ClientSecret, cfg.TokenURL)
}

func applyMatcherEnvFallbacks(cfg *matcher.Config) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = envOrDefault(envMatcherURL, "")
	}

	applyOAuthEnvFallbacks(
		&cfg.ClientID,
		&cfg.ClientSecret,
		&cfg.TokenURL,
		envMatcherClientID,
		envMatcherClientSecret,
		envMatcherTokenURL,
	)
}

func applyTracerEnvFallbacks(cfg *tracer.Config) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = envOrDefault(envTracerURL, "")
	}

	applyOAuthEnvFallbacks(
		&cfg.ClientID,
		&cfg.ClientSecret,
		&cfg.TokenURL,
		envTracerClientID,
		envTracerClientSecret,
		envTracerTokenURL,
	)
}

func applyReporterEnvFallbacks(cfg *reporter.Config) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = envOrDefault(envReporterURL, "")
	}

	if cfg.OrganizationID == "" {
		cfg.OrganizationID = envOrDefault(envReporterOrgID, "")
	}

	applyOAuthEnvFallbacks(
		&cfg.ClientID,
		&cfg.ClientSecret,
		&cfg.TokenURL,
		envReporterClientID,
		envReporterClientSecret,
		envReporterTokenURL,
	)
}

func applyFeesEnvFallbacks(cfg *fees.Config) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = envOrDefault(envFeesURL, "")
	}

	if cfg.OrganizationID == "" {
		cfg.OrganizationID = envOrDefault(envFeesOrgID, "")
	}

	applyOAuthEnvFallbacks(
		&cfg.ClientID,
		&cfg.ClientSecret,
		&cfg.TokenURL,
		envFeesClientID,
		envFeesClientSecret,
		envFeesTokenURL,
	)
}

// ---------------------------------------------------------------------------
// Internal: product initializers
// ---------------------------------------------------------------------------

// initMidaz applies deferred Midaz options, validates the config, creates
// two backends (onboarding + transaction), and constructs the Midaz client.
func (c *Client) initMidaz() error {
	if !c.midazRequested {
		return nil // Midaz not requested
	}

	var cfg midaz.Config

	for _, opt := range c.midazOpts {
		if err := opt(&cfg); err != nil {
			return fmt.Errorf("lerian: midaz option: %w", err)
		}
	}

	if err := validateLegacyAuthEnvUnset(
		"midaz",
		envMidazClientID+", "+envMidazClientSecret+", and "+envMidazTokenURL,
		envMidazLegacyAuthToken,
	); err != nil {
		return err
	}

	// Apply environment variable fallbacks for fields not set by options.
	applyMidazEnvFallbacks(&cfg)

	if err := validateMidazAuthConfig(cfg); err != nil {
		return err
	}

	// Validate required fields.
	if cfg.OnboardingURL == "" {
		return fmt.Errorf("lerian: midaz: OnboardingURL is required; " +
			"use midaz.WithOnboardingURL(\"https://your-server:3000/v1\") " +
			"(use http://localhost:3000/v1 for local development only)")
	}

	if cfg.TransactionURL == "" {
		return fmt.Errorf("lerian: midaz: TransactionURL is required; " +
			"use midaz.WithTransactionURL(\"https://your-server:3001/v1\") " +
			"(use http://localhost:3001/v1 for local development only)")
	}

	// Warn about insecure (non-localhost HTTP) URLs.
	warnInsecureURL("midaz (onboarding)", cfg.OnboardingURL)
	warnInsecureURL("midaz (transaction)", cfg.TransactionURL)

	authenticator := buildMidazAuthenticator(cfg, c.httpClientForTimeout(cfg.Timeout))

	// Create backends for the two Midaz microservices with the Midaz error parser.
	onboardingBackend := c.createBackend(cfg.OnboardingURL, authenticator, midaz.ParseError, nil, cfg.Timeout)
	transactionBackend := c.createBackend(cfg.TransactionURL, authenticator, midaz.ParseError, nil, cfg.Timeout)

	c.Midaz = midaz.NewClient(onboardingBackend, transactionBackend, cfg)

	return nil
}

// initMatcher applies deferred Matcher options, validates the config,
// creates a backend with OAuth2 client-credentials authentication, and constructs the Matcher client.
func (c *Client) initMatcher() error {
	if !c.matcherRequested {
		return nil // Matcher not requested
	}

	var cfg matcher.Config

	for _, opt := range c.matcherOpts {
		if err := opt(&cfg); err != nil {
			return fmt.Errorf("lerian: matcher option: %w", err)
		}
	}

	if err := validateLegacyAuthEnvUnset(
		"matcher",
		envMatcherClientID+", "+envMatcherClientSecret+", and "+envMatcherTokenURL,
		envMatcherLegacyAPIKey,
	); err != nil {
		return err
	}

	applyMatcherEnvFallbacks(&cfg)

	if err := validateOAuthAuthConfig("matcher", cfg.ClientID, cfg.ClientSecret, cfg.TokenURL); err != nil {
		return err
	}

	// Validate required fields.
	if cfg.BaseURL == "" {
		return fmt.Errorf("lerian: matcher: BaseURL is required; " +
			"use matcher.WithBaseURL(\"https://your-server:3002/v1\") " +
			"(use http://localhost:3002/v1 for local development only)")
	}

	// Warn about insecure (non-localhost HTTP) URLs.
	warnInsecureURL("matcher", cfg.BaseURL)

	authenticator := buildOAuthAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.TokenURL, cfg.Scopes, c.httpClientForTimeout(cfg.Timeout))

	backend := c.createBackend(cfg.BaseURL, authenticator, matcher.ParseError, nil, cfg.Timeout)
	c.Matcher = matcher.NewClient(backend, cfg)

	return nil
}

// initTracer applies deferred Tracer options, validates the config,
// creates a backend with OAuth2 client-credentials authentication, and constructs the Tracer client.
func (c *Client) initTracer() error {
	if !c.tracerRequested {
		return nil // Tracer not requested
	}

	var cfg tracer.Config

	for _, opt := range c.tracerOpts {
		if err := opt(&cfg); err != nil {
			return fmt.Errorf("lerian: tracer option: %w", err)
		}
	}

	if err := validateLegacyAuthEnvUnset(
		"tracer",
		envTracerClientID+", "+envTracerClientSecret+", and "+envTracerTokenURL,
		envTracerLegacyAPIKey,
	); err != nil {
		return err
	}

	applyTracerEnvFallbacks(&cfg)

	if err := validateOAuthAuthConfig("tracer", cfg.ClientID, cfg.ClientSecret, cfg.TokenURL); err != nil {
		return err
	}

	// Validate required fields.
	if cfg.BaseURL == "" {
		return fmt.Errorf("lerian: tracer: BaseURL is required; " +
			"use tracer.WithBaseURL(\"https://your-server:3003/v1\") " +
			"(use http://localhost:3003/v1 for local development only)")
	}

	// Warn about insecure (non-localhost HTTP) URLs.
	warnInsecureURL("tracer", cfg.BaseURL)

	authenticator := buildOAuthAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.TokenURL, cfg.Scopes, c.httpClientForTimeout(cfg.Timeout))

	backend := c.createBackend(cfg.BaseURL, authenticator, tracer.ParseError, nil, cfg.Timeout)
	c.Tracer = tracer.NewClient(backend, cfg)

	return nil
}

// initReporter applies deferred Reporter options, validates the config,
// creates a backend with OAuth2 client-credentials authentication and X-Organization-Id
// default header, and constructs the Reporter client.
func (c *Client) initReporter() error {
	if !c.reporterRequested {
		return nil // Reporter not requested
	}

	var cfg reporter.Config

	for _, opt := range c.reporterOpts {
		if err := opt(&cfg); err != nil {
			return fmt.Errorf("lerian: reporter option: %w", err)
		}
	}

	if err := validateLegacyAuthEnvUnset(
		"reporter",
		envReporterClientID+", "+envReporterClientSecret+", and "+envReporterTokenURL,
		envReporterLegacyAuthToken,
	); err != nil {
		return err
	}

	applyReporterEnvFallbacks(&cfg)

	if err := validateOAuthAuthConfig("reporter", cfg.ClientID, cfg.ClientSecret, cfg.TokenURL); err != nil {
		return err
	}

	// Validate required fields.
	if cfg.BaseURL == "" {
		return fmt.Errorf("lerian: reporter: BaseURL is required; " +
			"use reporter.WithBaseURL(\"https://your-server:3004/v1\") " +
			"(use http://localhost:3004/v1 for local development only)")
	}

	if cfg.OrganizationID == "" {
		return fmt.Errorf("lerian: reporter: OrganizationID is required; " +
			"use reporter.WithOrganizationID(\"org-uuid\")")
	}

	// Warn about insecure (non-localhost HTTP) URLs.
	warnInsecureURL("reporter", cfg.BaseURL)

	authenticator := buildOAuthAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.TokenURL, cfg.Scopes, c.httpClientForTimeout(cfg.Timeout))

	// Default headers include the organization scope.
	defaultHeaders := map[string]string{
		"X-Organization-Id": cfg.OrganizationID,
	}

	backend := c.createBackend(cfg.BaseURL, authenticator, reporter.ParseError, defaultHeaders, cfg.Timeout)
	c.Reporter = reporter.NewClient(backend, cfg)

	return nil
}

// initFees applies deferred Fees options, validates the config, creates
// a backend with OAuth2 client-credentials authentication and X-Organization-Id default
// header, and constructs the Fees client.
func (c *Client) initFees() error {
	if !c.feesRequested {
		return nil // Fees not requested
	}

	var cfg fees.Config

	for _, opt := range c.feesOpts {
		if err := opt(&cfg); err != nil {
			return fmt.Errorf("lerian: fees option: %w", err)
		}
	}

	if err := validateLegacyAuthEnvUnset(
		"fees",
		envFeesClientID+", "+envFeesClientSecret+", and "+envFeesTokenURL,
		envFeesLegacyAuthToken,
	); err != nil {
		return err
	}

	applyFeesEnvFallbacks(&cfg)

	if err := validateOAuthAuthConfig("fees", cfg.ClientID, cfg.ClientSecret, cfg.TokenURL); err != nil {
		return err
	}

	// Validate required fields.
	if cfg.BaseURL == "" {
		return fmt.Errorf("lerian: fees: BaseURL is required; " +
			"use fees.WithBaseURL(\"https://your-server:3005/v1\") " +
			"(use http://localhost:3005/v1 for local development only)")
	}

	if cfg.OrganizationID == "" {
		return fmt.Errorf("lerian: fees: OrganizationID is required; " +
			"use fees.WithOrganizationID(\"org-uuid\")")
	}

	// Warn about insecure (non-localhost HTTP) URLs.
	warnInsecureURL("fees", cfg.BaseURL)

	authenticator := buildOAuthAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.TokenURL, cfg.Scopes, c.httpClientForTimeout(cfg.Timeout))

	// Default headers include the organization scope.
	defaultHeaders := map[string]string{
		"X-Organization-Id": cfg.OrganizationID,
	}

	backend := c.createBackend(cfg.BaseURL, authenticator, fees.ParseError, defaultHeaders, cfg.Timeout)
	c.Fees = fees.NewClient(backend, cfg)

	return nil
}
