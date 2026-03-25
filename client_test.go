package lerian

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/retry"
	"github.com/LerianStudio/lerian-sdk-golang/reporter"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustMidazConfig(t *testing.T, opts ...midaz.Option) *midaz.Config {
	t.Helper()

	var cfg midaz.Config
	for _, opt := range opts {
		require.NoError(t, opt(&cfg))
	}

	return &cfg
}

func mustMatcherConfig(t *testing.T, opts ...matcher.Option) *matcher.Config {
	t.Helper()

	var cfg matcher.Config
	for _, opt := range opts {
		require.NoError(t, opt(&cfg))
	}

	return &cfg
}

func mustTracerConfig(t *testing.T, opts ...tracer.Option) *tracer.Config {
	t.Helper()

	var cfg tracer.Config
	for _, opt := range opts {
		require.NoError(t, opt(&cfg))
	}

	return &cfg
}

func mustReporterConfig(t *testing.T, opts ...reporter.Option) *reporter.Config {
	t.Helper()

	var cfg reporter.Config
	for _, opt := range opts {
		require.NoError(t, opt(&cfg))
	}

	return &cfg
}

func mustFeesConfig(t *testing.T, opts ...fees.Option) *fees.Config {
	t.Helper()

	var cfg fees.Config
	for _, opt := range opts {
		require.NoError(t, opt(&cfg))
	}

	return &cfg
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func assertLerianOAuthPayload(t *testing.T, payload map[string]string, clientID, clientSecret string) {
	t.Helper()

	assert.Equal(t, map[string]string{
		"grantType":    "client_credentials",
		"clientId":     clientID,
		"clientSecret": clientSecret,
	}, payload)
}

func tokenResponseHTTPResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestBuildOAuthAuthenticatorUsesProvidedHTTPClient(t *testing.T) {
	t.Parallel()

	requestCount := 0
	customClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requestCount++

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)

			var payload map[string]string
			require.NoError(t, json.Unmarshal(body, &payload))
			assertLerianOAuthPayload(t, payload, "cid", "csecret")

			return tokenResponseHTTPResponse(`{"accessToken":"tok-built","tokenType":"Bearer","expiresIn":3600,"refreshToken":"refresh-built"}`), nil
		}),
	}

	authenticator := buildOAuthAuthenticator("cid", "csecret", "https://auth.example.com/token", customClient)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	require.NoError(t, authenticator.Enrich(context.Background(), req))
	assert.Equal(t, "Bearer tok-built", req.Header.Get("Authorization"))
	assert.Equal(t, 1, requestCount)
}

func TestBuildOAuthAuthenticatorReturnsNoAuthWithoutCredentials(t *testing.T) {
	t.Parallel()

	authenticator := buildOAuthAuthenticator("", "", "", nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	require.NoError(t, authenticator.Enrich(context.Background(), req))
	assert.Empty(t, req.Header.Get("Authorization"))
}

func TestHTTPClientForTimeoutClonesConfiguredClient(t *testing.T) {
	t.Parallel()

	base := &http.Client{Timeout: defaultHTTPTimeout}
	c := &Client{httpClient: base}

	cloned := c.httpClientForTimeout(42 * time.Second)

	require.NotNil(t, cloned)
	assert.NotSame(t, base, cloned)
	assert.Equal(t, 42*time.Second, cloned.Timeout)
}

func TestValidateOAuthAuthConfigRejectsRemoteHTTPTokenURL(t *testing.T) {
	t.Parallel()

	err := validateOAuthAuthConfig("matcher", "cid", "secret", "http://auth.example.com/token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TokenURL must use HTTPS outside localhost")
}

func TestValidateOAuthAuthConfigAllowsLocalHTTPTokenURL(t *testing.T) {
	t.Parallel()

	err := validateOAuthAuthConfig("matcher", "cid", "secret", "http://localhost:8080/token")
	require.NoError(t, err)
}

func TestNewEmptyClient(t *testing.T) {
	t.Parallel()

	client, err := New(Config{})
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.Nil(t, client.Midaz)
	assert.Nil(t, client.Matcher)
	assert.Nil(t, client.Tracer)
	assert.Nil(t, client.Reporter)
	assert.Nil(t, client.Fees)
	assert.NotNil(t, client.observability)
	assert.False(t, client.observability.IsEnabled())
	assert.Equal(t, defaultHTTPTimeout, client.httpClient.Timeout)
	assert.Equal(t, retry.DefaultConfig(), client.retryConfig)
}

func TestNewWithMidaz(t *testing.T) {
	t.Parallel()

	client, err := New(Config{
		Midaz: mustMidazConfig(t,
			midaz.WithOnboardingURL("http://localhost:3000/v1"),
			midaz.WithTransactionURL("http://localhost:3001/v1"),
			midaz.WithClientCredentials("client-id", "client-secret", "http://localhost:8080/token"),
		),
	})
	require.NoError(t, err)
	require.NotNil(t, client.Midaz)
	assert.NotNil(t, client.Midaz.CRM)
	assert.NotNil(t, client.Midaz.CRM.Holders)
	assert.NotNil(t, client.Midaz.CRM.Aliases)
	assert.Nil(t, client.Matcher)
}

func TestNewWithMultipleProducts(t *testing.T) {
	t.Parallel()

	client, err := New(Config{
		Midaz: mustMidazConfig(t,
			midaz.WithOnboardingURL("http://localhost:3000/v1"),
			midaz.WithTransactionURL("http://localhost:3001/v1"),
		),
		Matcher: mustMatcherConfig(t, matcher.WithBaseURL("http://localhost:3002/v1")),
		Tracer:  mustTracerConfig(t, tracer.WithBaseURL("http://localhost:3003/v1")),
		Reporter: mustReporterConfig(t,
			reporter.WithBaseURL("http://localhost:3004/v1"),
			reporter.WithOrganizationID("org-1"),
		),
		Fees: mustFeesConfig(t,
			fees.WithBaseURL("http://localhost:3005/v1"),
			fees.WithOrganizationID("org-1"),
		),
	})
	require.NoError(t, err)

	assert.NotNil(t, client.Midaz)
	assert.NotNil(t, client.Matcher)
	assert.NotNil(t, client.Tracer)
	assert.NotNil(t, client.Reporter)
	assert.NotNil(t, client.Fees)
}

func TestMidazCreatesOptionalCRMBackend(t *testing.T) {
	t.Parallel()

	var crmRequests atomic.Int32

	crmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		crmRequests.Add(1)
		assert.Equal(t, "/holders", r.URL.Path)
		assert.Equal(t, "org-1", r.Header.Get("X-Organization-Id"))

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"items":[{"id":"holder-1","name":"Acme Holder"}]}`))
		require.NoError(t, err)
	}))
	defer crmServer.Close()

	stubServer := httptest.NewServer(http.NotFoundHandler())
	defer stubServer.Close()

	client, err := New(Config{
		Midaz: &midaz.Config{
			OnboardingURL:  stubServer.URL,
			TransactionURL: stubServer.URL,
			CRMURL:         crmServer.URL,
		},
	})
	require.NoError(t, err)

	iter := client.Midaz.CRM.Holders.List(context.Background(), "org-1", nil)
	holders, err := iter.Collect(context.Background())
	require.NoError(t, err)
	require.Len(t, holders, 1)
	assert.Equal(t, int32(1), crmRequests.Load())
}

func TestMidazNoAuthAppliesToBothBackends(t *testing.T) {
	t.Parallel()

	var (
		onboardingAuth  string
		transactionAuth string
	)

	onboardingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		onboardingAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"id":"org-1"}`))
		require.NoError(t, err)
	}))
	defer onboardingServer.Close()

	transactionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transactionAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"id":"tx-1"}`))
		require.NoError(t, err)
	}))
	defer transactionServer.Close()

	client, err := New(Config{
		Midaz: &midaz.Config{
			OnboardingURL:  onboardingServer.URL,
			TransactionURL: transactionServer.URL,
		},
	})
	require.NoError(t, err)

	_, err = client.Midaz.Onboarding.Organizations.Get(context.Background(), "org-1")
	require.NoError(t, err)
	_, err = client.Midaz.Transactions.Transactions.Get(context.Background(), "org-1", "ledger-1", "tx-1")
	require.NoError(t, err)

	assert.Empty(t, onboardingAuth)
	assert.Empty(t, transactionAuth)
}

func TestNewUsesCustomHTTPClient(t *testing.T) {
	t.Parallel()

	custom := &http.Client{Timeout: 5 * time.Second}
	client, err := New(Config{HTTPClient: custom})
	require.NoError(t, err)
	assert.Same(t, custom, client.httpClient)
}

func TestNewUsesCustomRetryConfig(t *testing.T) {
	t.Parallel()

	custom := &retry.Config{MaxRetries: 0, BaseDelay: 0}
	client, err := New(Config{RetryConfig: custom})
	require.NoError(t, err)
	assert.Equal(t, *custom, client.retryConfig)
}

func TestNewUsesDebugFlag(t *testing.T) {
	t.Parallel()

	client, err := New(Config{Debug: true})
	require.NoError(t, err)
	assert.True(t, client.debug)
}

func TestPerProductTimeoutCreatesClonedHTTPClient(t *testing.T) {
	t.Parallel()

	client, err := New(Config{
		Matcher: mustMatcherConfig(t,
			matcher.WithBaseURL("http://localhost:3002/v1"),
			matcher.WithTimeout(42*time.Second),
		),
	})
	require.NoError(t, err)
	assert.Equal(t, defaultHTTPTimeout, client.httpClient.Timeout)
	assert.NotNil(t, client.Matcher)
}

func TestNewInvalidMidazMissingOnboardingURL(t *testing.T) {
	t.Parallel()

	_, err := New(Config{Midaz: mustMidazConfig(t, midaz.WithTransactionURL("http://localhost:3001/v1"))})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OnboardingURL is required")
	assert.Contains(t, err.Error(), "Config.Midaz.OnboardingURL")
}

func TestNewInvalidMidazMissingTransactionURL(t *testing.T) {
	t.Parallel()

	_, err := New(Config{Midaz: mustMidazConfig(t, midaz.WithOnboardingURL("http://localhost:3000/v1"))})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TransactionURL is required")
	assert.Contains(t, err.Error(), "Config.Midaz.TransactionURL")
}

func TestNewInvalidMatcherMissingBaseURL(t *testing.T) {
	t.Parallel()

	_, err := New(Config{Matcher: &matcher.Config{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matcher: BaseURL is required")
}

func TestNewInvalidTracerMissingBaseURL(t *testing.T) {
	t.Parallel()

	_, err := New(Config{Tracer: &tracer.Config{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tracer: BaseURL is required")
}

func TestNewInvalidReporterMissingRequiredFields(t *testing.T) {
	t.Parallel()

	_, err := New(Config{Reporter: &reporter.Config{OrganizationID: "org-123"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reporter: BaseURL is required")

	_, err = New(Config{Reporter: &reporter.Config{BaseURL: "http://localhost:3004/v1"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OrganizationID is required")
	assert.Contains(t, err.Error(), "Config.Reporter.OrganizationID")
}

func TestNewInvalidFeesMissingRequiredFields(t *testing.T) {
	t.Parallel()

	_, err := New(Config{Fees: &fees.Config{OrganizationID: "org-123"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fees: BaseURL is required")

	_, err = New(Config{Fees: &fees.Config{BaseURL: "http://localhost:3005/v1"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OrganizationID is required")
	assert.Contains(t, err.Error(), "Config.Fees.OrganizationID")
}

func TestShutdownIdempotency(t *testing.T) {
	t.Parallel()

	client, err := New(Config{})
	require.NoError(t, err)

	err = client.Shutdown(context.Background())
	assert.NoError(t, err)
	err = client.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestShutdownNilObservability(t *testing.T) {
	t.Parallel()

	c := &Client{}
	err := c.Shutdown(context.Background())
	assert.NoError(t, err)
}

func captureSlogOutput(fn func()) string {
	var buf bytes.Buffer

	original := slog.Default()
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})

	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(original)

	fn()

	return buf.String()
}

func TestIsLocalhostURL(t *testing.T) {
	t.Parallel()

	assert.True(t, isLocalhostURL("http://localhost:3000/v1"))
	assert.True(t, isLocalhostURL("http://127.0.0.1:3000/v1"))
	assert.True(t, isLocalhostURL("http://[::1]:3000/v1"))
	assert.False(t, isLocalhostURL("http://api.example.com/v1"))
	assert.False(t, isLocalhostURL(""))
}

func TestWarnInsecureURL(t *testing.T) {
	t.Parallel()

	output := captureSlogOutput(func() {
		warnInsecureURL("midaz", "http://api.example.com:3000/v1")
	})
	assert.Contains(t, output, "insecure URL detected")
	assert.Contains(t, output, "midaz")

	output = captureSlogOutput(func() {
		warnInsecureURL("midaz", "http://localhost:3000/v1")
	})
	assert.Empty(t, output)
}
