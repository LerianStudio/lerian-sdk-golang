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
// client.Matcher). Products with nil config in [Config] remain nil.
type Client struct {
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
	shutdownOnce  sync.Once
}

// New constructs a fully-wired [Client] from an explicit root [Config].
// Use [LoadConfigFromEnv] when you want to build that config from `LERIAN_*`
// environment variables before construction.
func New(cfg Config) (*Client, error) {
	retryConfig := retry.DefaultConfig()
	if cfg.RetryConfig != nil {
		retryConfig = *cfg.RetryConfig
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}

	c := &Client{
		retryConfig: retryConfig,
		jsonPool:    performance.NewJSONPool(),
		httpClient:  httpClient,
		debug:       cfg.Debug,
	}

	if cfg.Observability.Traces || cfg.Observability.Metrics || cfg.Observability.Logs {
		provider, err := observability.NewProvider(observability.ProviderConfig{
			ServiceName:       "lerian-sdk-go",
			ServiceVersion:    sdkVersion,
			CollectorEndpoint: cfg.Observability.CollectorEndpoint,
			EnableTraces:      cfg.Observability.Traces,
			EnableMetrics:     cfg.Observability.Metrics,
			EnableLogs:        cfg.Observability.Logs,
		})
		if err != nil {
			return nil, fmt.Errorf("lerian: initializing observability: %w", err)
		}

		c.observability = provider
	} else {
		c.observability = observability.NewNoopProvider()
	}

	if err := c.initMidaz(cfg.Midaz); err != nil {
		return nil, err
	}

	if err := c.initMatcher(cfg.Matcher); err != nil {
		return nil, err
	}

	if err := c.initTracer(cfg.Tracer); err != nil {
		return nil, err
	}

	if err := c.initReporter(cfg.Reporter); err != nil {
		return nil, err
	}

	if err := c.initFees(cfg.Fees); err != nil {
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

func buildOAuthAuthenticator(clientID, clientSecret, tokenURL string, httpClient *http.Client) auth.Authenticator {
	if clientID != "" && clientSecret != "" && tokenURL != "" {
		return auth.NewOAuth2WithHTTPClient(clientID, clientSecret, tokenURL, httpClient)
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

// ---------------------------------------------------------------------------
// Internal: product initializers
// ---------------------------------------------------------------------------

// initMidaz validates the config, creates the required
// onboarding/transaction backends plus an optional CRM backend, and constructs
// the Midaz client.
func (c *Client) initMidaz(cfg *midaz.Config) error {
	if cfg == nil {
		return nil
	}

	resolved := *cfg

	if err := validateLegacyAuthEnvUnset(
		"midaz",
		envMidazClientID+", "+envMidazClientSecret+", and "+envMidazTokenURL,
		envMidazLegacyAuthToken,
	); err != nil {
		return err
	}

	if err := validateOAuthAuthConfig("midaz", resolved.ClientID, resolved.ClientSecret, resolved.TokenURL); err != nil {
		return err
	}

	if resolved.OnboardingURL == "" {
		return fmt.Errorf("lerian: midaz: OnboardingURL is required; " +
			"set Config.Midaz.OnboardingURL to \"https://your-server:3000/v1\" " +
			"(use http://localhost:3000/v1 for local development only)")
	}

	if resolved.TransactionURL == "" {
		return fmt.Errorf("lerian: midaz: TransactionURL is required; " +
			"set Config.Midaz.TransactionURL to \"https://your-server:3001/v1\" " +
			"(use http://localhost:3001/v1 for local development only)")
	}

	warnInsecureURL("midaz (onboarding)", resolved.OnboardingURL)
	warnInsecureURL("midaz (transaction)", resolved.TransactionURL)

	if resolved.CRMURL != "" {
		warnInsecureURL("midaz (crm)", resolved.CRMURL)
	}

	authenticator := buildOAuthAuthenticator(resolved.ClientID, resolved.ClientSecret, resolved.TokenURL, c.httpClientForTimeout(resolved.Timeout))

	onboardingBackend := c.createBackend(resolved.OnboardingURL, authenticator, midaz.ParseError, nil, resolved.Timeout)
	transactionBackend := c.createBackend(resolved.TransactionURL, authenticator, midaz.ParseError, nil, resolved.Timeout)

	var crmBackend core.Backend
	if resolved.CRMURL != "" {
		crmBackend = c.createBackend(resolved.CRMURL, authenticator, midaz.ParseError, nil, resolved.Timeout)
	}

	c.Midaz = midaz.NewClientWithCRM(onboardingBackend, transactionBackend, crmBackend, resolved)

	return nil
}

func (c *Client) initMatcher(cfg *matcher.Config) error {
	if cfg == nil {
		return nil
	}

	resolved := *cfg

	if err := validateLegacyAuthEnvUnset(
		"matcher",
		envMatcherClientID+", "+envMatcherClientSecret+", and "+envMatcherTokenURL,
		envMatcherLegacyAPIKey,
	); err != nil {
		return err
	}

	if err := validateOAuthAuthConfig("matcher", resolved.ClientID, resolved.ClientSecret, resolved.TokenURL); err != nil {
		return err
	}

	if resolved.BaseURL == "" {
		return fmt.Errorf("lerian: matcher: BaseURL is required; " +
			"set Config.Matcher.BaseURL to \"https://your-server:3002/v1\" " +
			"(use http://localhost:3002/v1 for local development only)")
	}

	warnInsecureURL("matcher", resolved.BaseURL)

	authenticator := buildOAuthAuthenticator(resolved.ClientID, resolved.ClientSecret, resolved.TokenURL, c.httpClientForTimeout(resolved.Timeout))

	backend := c.createBackend(resolved.BaseURL, authenticator, matcher.ParseError, nil, resolved.Timeout)
	c.Matcher = matcher.NewClient(backend, resolved)

	return nil
}

func (c *Client) initTracer(cfg *tracer.Config) error {
	if cfg == nil {
		return nil
	}

	resolved := *cfg

	if err := validateLegacyAuthEnvUnset(
		"tracer",
		envTracerClientID+", "+envTracerClientSecret+", and "+envTracerTokenURL,
		envTracerLegacyAPIKey,
	); err != nil {
		return err
	}

	if err := validateOAuthAuthConfig("tracer", resolved.ClientID, resolved.ClientSecret, resolved.TokenURL); err != nil {
		return err
	}

	if resolved.BaseURL == "" {
		return fmt.Errorf("lerian: tracer: BaseURL is required; " +
			"set Config.Tracer.BaseURL to \"https://your-server:3003/v1\" " +
			"(use http://localhost:3003/v1 for local development only)")
	}

	warnInsecureURL("tracer", resolved.BaseURL)

	authenticator := buildOAuthAuthenticator(resolved.ClientID, resolved.ClientSecret, resolved.TokenURL, c.httpClientForTimeout(resolved.Timeout))

	backend := c.createBackend(resolved.BaseURL, authenticator, tracer.ParseError, nil, resolved.Timeout)
	c.Tracer = tracer.NewClient(backend, resolved)

	return nil
}

func (c *Client) initReporter(cfg *reporter.Config) error {
	if cfg == nil {
		return nil
	}

	resolved := *cfg

	if err := validateLegacyAuthEnvUnset(
		"reporter",
		envReporterClientID+", "+envReporterClientSecret+", and "+envReporterTokenURL,
		envReporterLegacyAuthToken,
	); err != nil {
		return err
	}

	if err := validateOAuthAuthConfig("reporter", resolved.ClientID, resolved.ClientSecret, resolved.TokenURL); err != nil {
		return err
	}

	if resolved.BaseURL == "" {
		return fmt.Errorf("lerian: reporter: BaseURL is required; " +
			"set Config.Reporter.BaseURL to \"https://your-server:3004/v1\" " +
			"(use http://localhost:3004/v1 for local development only)")
	}

	if resolved.OrganizationID == "" {
		return fmt.Errorf("lerian: reporter: OrganizationID is required; " +
			"set Config.Reporter.OrganizationID to \"org-uuid\"")
	}

	warnInsecureURL("reporter", resolved.BaseURL)

	authenticator := buildOAuthAuthenticator(resolved.ClientID, resolved.ClientSecret, resolved.TokenURL, c.httpClientForTimeout(resolved.Timeout))

	defaultHeaders := map[string]string{
		"X-Organization-Id": resolved.OrganizationID,
	}

	backend := c.createBackend(resolved.BaseURL, authenticator, reporter.ParseError, defaultHeaders, resolved.Timeout)
	c.Reporter = reporter.NewClient(backend, resolved)

	return nil
}

func (c *Client) initFees(cfg *fees.Config) error {
	if cfg == nil {
		return nil
	}

	resolved := *cfg

	if err := validateLegacyAuthEnvUnset(
		"fees",
		envFeesClientID+", "+envFeesClientSecret+", and "+envFeesTokenURL,
		envFeesLegacyAuthToken,
	); err != nil {
		return err
	}

	if err := validateOAuthAuthConfig("fees", resolved.ClientID, resolved.ClientSecret, resolved.TokenURL); err != nil {
		return err
	}

	if resolved.BaseURL == "" {
		return fmt.Errorf("lerian: fees: BaseURL is required; " +
			"set Config.Fees.BaseURL to \"https://your-server:3005/v1\" " +
			"(use http://localhost:3005/v1 for local development only)")
	}

	if resolved.OrganizationID == "" {
		return fmt.Errorf("lerian: fees: OrganizationID is required; " +
			"set Config.Fees.OrganizationID to \"org-uuid\"")
	}

	warnInsecureURL("fees", resolved.BaseURL)

	authenticator := buildOAuthAuthenticator(resolved.ClientID, resolved.ClientSecret, resolved.TokenURL, c.httpClientForTimeout(resolved.Timeout))

	defaultHeaders := map[string]string{
		"X-Organization-Id": resolved.OrganizationID,
	}

	backend := c.createBackend(resolved.BaseURL, authenticator, fees.ParseError, defaultHeaders, resolved.Timeout)
	c.Fees = fees.NewClient(backend, resolved)

	return nil
}
