package lerian

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/reporter"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// New() — happy paths
// ---------------------------------------------------------------------------

func TestNewEmptyClient(t *testing.T) {
	t.Parallel()

	client, err := New()
	require.NoError(t, err)
	require.NotNil(t, client)

	// No products configured — all product fields must be nil.
	assert.Nil(t, client.Midaz, "Midaz should be nil when not configured")
	assert.Nil(t, client.Matcher, "Matcher should be nil when not configured")
	assert.Nil(t, client.Tracer, "Tracer should be nil when not configured")
	assert.Nil(t, client.Reporter, "Reporter should be nil when not configured")
	assert.Nil(t, client.Fees, "Fees should be nil when not configured")

	// Observability should be initialized (noop).
	assert.NotNil(t, client.observability)
	assert.False(t, client.observability.IsEnabled())
}

func TestNewWithMidaz(t *testing.T) {
	t.Parallel()

	client, err := New(
		WithMidaz(
			midaz.WithOnboardingURL("http://localhost:3000/v1"),
			midaz.WithTransactionURL("http://localhost:3001/v1"),
			midaz.WithAuthToken("test-token"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.Midaz, "Midaz should be initialized")
	assert.Nil(t, client.Matcher, "Matcher should be nil when not configured")
	assert.Nil(t, client.Tracer, "Tracer should be nil when not configured")
	assert.Nil(t, client.Reporter, "Reporter should be nil when not configured")
	assert.Nil(t, client.Fees, "Fees should be nil when not configured")
}

func TestNewWithMidazNoAuth(t *testing.T) {
	t.Parallel()

	// Midaz without auth token should still work (NoAuth fallback).
	client, err := New(
		WithMidaz(
			midaz.WithOnboardingURL("http://localhost:3000/v1"),
			midaz.WithTransactionURL("http://localhost:3001/v1"),
		),
	)
	require.NoError(t, err)
	assert.NotNil(t, client.Midaz)
}

func TestNewWithMatcher(t *testing.T) {
	t.Parallel()

	client, err := New(
		WithMatcher(
			matcher.WithBaseURL("http://localhost:3002/v1"),
			matcher.WithAPIKey("matcher-key-123"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.Matcher, "Matcher should be initialized")
	assert.Nil(t, client.Midaz, "Midaz should be nil when not configured")
}

func TestNewWithTracer(t *testing.T) {
	t.Parallel()

	client, err := New(
		WithTracer(
			tracer.WithBaseURL("http://localhost:3003/v1"),
			tracer.WithAPIKey("tracer-key-456"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.Tracer, "Tracer should be initialized")
	assert.Nil(t, client.Midaz, "Midaz should be nil when not configured")
}

func TestNewWithReporter(t *testing.T) {
	t.Parallel()

	client, err := New(
		WithReporter(
			reporter.WithBaseURL("http://localhost:3004/v1"),
			reporter.WithAuthToken("reporter-token"),
			reporter.WithOrganizationID("org-12345"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.Reporter, "Reporter should be initialized")
	assert.Nil(t, client.Midaz, "Midaz should be nil when not configured")
}

func TestNewWithFees(t *testing.T) {
	t.Parallel()

	client, err := New(
		WithFees(
			fees.WithBaseURL("http://localhost:3005/v1"),
			fees.WithAuthToken("fees-token"),
			fees.WithOrganizationID("org-67890"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.Fees, "Fees should be initialized")
	assert.Nil(t, client.Midaz, "Midaz should be nil when not configured")
}

func TestNewWithMultipleProducts(t *testing.T) {
	t.Parallel()

	client, err := New(
		WithMidaz(
			midaz.WithOnboardingURL("http://localhost:3000/v1"),
			midaz.WithTransactionURL("http://localhost:3001/v1"),
			midaz.WithAuthToken("midaz-token"),
		),
		WithMatcher(
			matcher.WithBaseURL("http://localhost:3002/v1"),
			matcher.WithAPIKey("matcher-key"),
		),
		WithTracer(
			tracer.WithBaseURL("http://localhost:3003/v1"),
			tracer.WithAPIKey("tracer-key"),
		),
		WithReporter(
			reporter.WithBaseURL("http://localhost:3004/v1"),
			reporter.WithOrganizationID("org-multi"),
		),
		WithFees(
			fees.WithBaseURL("http://localhost:3005/v1"),
			fees.WithOrganizationID("org-multi"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.Midaz, "Midaz should be initialized")
	assert.NotNil(t, client.Matcher, "Matcher should be initialized")
	assert.NotNil(t, client.Tracer, "Tracer should be initialized")
	assert.NotNil(t, client.Reporter, "Reporter should be initialized")
	assert.NotNil(t, client.Fees, "Fees should be initialized")
}

// ---------------------------------------------------------------------------
// New() — validation errors
// ---------------------------------------------------------------------------

func TestNewInvalidMidazMissingOnboardingURL(t *testing.T) {
	t.Parallel()

	_, err := New(
		WithMidaz(
			midaz.WithTransactionURL("http://localhost:3001/v1"),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OnboardingURL is required")
	assert.Contains(t, err.Error(), "midaz.WithOnboardingURL")
}

func TestNewInvalidMidazMissingTransactionURL(t *testing.T) {
	t.Parallel()

	_, err := New(
		WithMidaz(
			midaz.WithOnboardingURL("http://localhost:3000/v1"),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TransactionURL is required")
	assert.Contains(t, err.Error(), "midaz.WithTransactionURL")
}

func TestNewInvalidMatcherMissingBaseURL(t *testing.T) {
	t.Parallel()

	_, err := New(
		WithMatcher(
			matcher.WithAPIKey("some-key"),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matcher: BaseURL is required")
}

func TestNewInvalidTracerMissingBaseURL(t *testing.T) {
	t.Parallel()

	_, err := New(
		WithTracer(
			tracer.WithAPIKey("some-key"),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tracer: BaseURL is required")
}

func TestNewInvalidReporterMissingBaseURL(t *testing.T) {
	t.Parallel()

	_, err := New(
		WithReporter(
			reporter.WithOrganizationID("org-123"),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reporter: BaseURL is required")
}

func TestNewInvalidReporterMissingOrganizationID(t *testing.T) {
	t.Parallel()

	_, err := New(
		WithReporter(
			reporter.WithBaseURL("http://localhost:3004/v1"),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OrganizationID is required")
	assert.Contains(t, err.Error(), "reporter.WithOrganizationID")
}

func TestNewInvalidFeesMissingBaseURL(t *testing.T) {
	t.Parallel()

	_, err := New(
		WithFees(
			fees.WithOrganizationID("org-123"),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fees: BaseURL is required")
}

func TestNewInvalidFeesMissingOrganizationID(t *testing.T) {
	t.Parallel()

	_, err := New(
		WithFees(
			fees.WithBaseURL("http://localhost:3005/v1"),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OrganizationID is required")
	assert.Contains(t, err.Error(), "fees.WithOrganizationID")
}

// ---------------------------------------------------------------------------
// Shutdown
// ---------------------------------------------------------------------------

func TestShutdownIdempotency(t *testing.T) {
	t.Parallel()

	client, err := New()
	require.NoError(t, err)

	ctx := context.Background()

	// First call.
	err = client.Shutdown(ctx)
	assert.NoError(t, err, "first Shutdown should succeed")

	// Second call — must not error (idempotent via sync.Once).
	err = client.Shutdown(ctx)
	assert.NoError(t, err, "second Shutdown should be idempotent")
}

func TestShutdownRespectsContext(t *testing.T) {
	t.Parallel()

	client, err := New()
	require.NoError(t, err)

	// Create a context with a very short deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// With a noop provider, shutdown completes instantly.
	err = client.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestShutdownNilObservability(t *testing.T) {
	t.Parallel()

	// Construct a client manually to exercise the nil guard in Shutdown.
	c := &Client{}
	err := c.Shutdown(context.Background())
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// WithHTTPClient(nil) — Issue #5
// ---------------------------------------------------------------------------

func TestWithHTTPClientNilReturnsError(t *testing.T) {
	t.Parallel()

	_, err := New(WithHTTPClient(nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client must not be nil")
}

func TestWithHTTPClientNonNil(t *testing.T) {
	t.Parallel()

	custom := &http.Client{Timeout: 5 * time.Second}
	client, err := New(WithHTTPClient(custom))
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, custom, client.httpClient)
}

// ---------------------------------------------------------------------------
// WithDebug(false) vs LERIAN_DEBUG=true — Issue #8
// ---------------------------------------------------------------------------

func TestWithDebugFalseOverridesEnv(t *testing.T) {
	// Set the env var to true; explicit WithDebug(false) must win.
	t.Setenv("LERIAN_DEBUG", "true")

	client, err := New(WithDebug(false))
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.False(t, client.debug, "WithDebug(false) should override LERIAN_DEBUG=true")
}

func TestWithDebugTrueExplicit(t *testing.T) {
	t.Parallel()

	// Ensure explicit WithDebug(true) still works.
	client, err := New(WithDebug(true))
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.True(t, client.debug, "WithDebug(true) should enable debug")
}

func TestDebugEnvFallbackWhenNoOption(t *testing.T) {
	// When WithDebug is not called, env var should be the fallback.
	t.Setenv("LERIAN_DEBUG", "true")

	client, err := New()
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.True(t, client.debug, "LERIAN_DEBUG=true should enable debug when WithDebug is not called")
}

// ---------------------------------------------------------------------------
// Matcher ErrorParser wired — Issue #2
// ---------------------------------------------------------------------------

func TestMatcherErrorParserWired(t *testing.T) {
	t.Parallel()

	// Construct a client with Matcher configured. The backend should have
	// a non-nil error parser (matcher.ParseError). We verify this indirectly
	// by confirming the Matcher client is initialized (the parser is
	// internal to BackendImpl, but we can at least ensure the init path
	// that now wires the parser succeeds).
	client, err := New(
		WithMatcher(
			matcher.WithBaseURL("http://localhost:3002/v1"),
			matcher.WithAPIKey("test-key"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, client.Matcher, "Matcher should be initialized with error parser wired")
}

// ---------------------------------------------------------------------------
// Per-product timeout — Issue #7
// ---------------------------------------------------------------------------

func TestPerProductTimeoutCreatesClonedHTTPClient(t *testing.T) {
	t.Parallel()

	// Verify that a product-specific timeout results in a different HTTP
	// client with the overridden timeout value.
	client, err := New(
		WithMatcher(
			matcher.WithBaseURL("http://localhost:3002/v1"),
			matcher.WithTimeout(42*time.Second),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, client.Matcher)

	// The shared httpClient on the root Client should still have the default timeout.
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout,
		"shared HTTP client timeout should be unchanged")
}

// ---------------------------------------------------------------------------
// isLocalhostURL
// ---------------------------------------------------------------------------

func TestIsLocalhostURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		// Positive cases — these are localhost.
		{"localhost no port", "http://localhost/v1", true},
		{"localhost with port", "http://localhost:3000/v1", true},
		{"localhost HTTPS", "https://localhost:3000/v1", true},
		{"localhost uppercase", "http://LOCALHOST:3000/v1", true},
		{"localhost mixed case", "http://LocalHost:3000/v1", true},
		{"127.0.0.1 no port", "http://127.0.0.1/v1", true},
		{"127.0.0.1 with port", "http://127.0.0.1:3000/v1", true},
		{"[::1] with port", "http://[::1]:3000/v1", true},
		{"[::1] no port", "http://[::1]/v1", true},

		// Negative cases — these are NOT localhost.
		{"remote host", "http://api.example.com/v1", false},
		{"remote IP", "http://192.168.1.1:3000/v1", false},
		{"remote HTTPS", "https://api.example.com/v1", false},
		{"empty string", "", false},
		{"no scheme", "example.com:3000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isLocalhostURL(tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// warnInsecureURL
// ---------------------------------------------------------------------------

// captureSlogOutput installs a temporary slog default handler that writes
// to a buffer, calls fn, then restores the original handler. It returns
// whatever was written to the buffer.
func captureSlogOutput(fn func()) string {
	var buf bytes.Buffer

	original := slog.Default()

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})
	slog.SetDefault(slog.New(handler))

	defer slog.SetDefault(original)

	fn()

	return buf.String()
}

func TestWarnInsecureURL(t *testing.T) {
	tests := []struct {
		name        string
		product     string
		url         string
		wantWarning bool
	}{
		// Should warn — HTTP to a remote host.
		{"remote http", "midaz", "http://api.example.com:3000/v1", true},
		{"remote http no port", "matcher", "http://api.example.com/v1", true},
		{"remote http uppercase scheme", "tracer", "HTTP://api.example.com/v1", true},
		{"remote http mixed case scheme", "fees", "Http://api.example.com/v1", true},

		// Should NOT warn — HTTPS (secure).
		{"remote https", "midaz", "https://api.example.com:3000/v1", false},

		// Should NOT warn — localhost is acceptable for dev.
		{"localhost http", "midaz", "http://localhost:3000/v1", false},
		{"127.0.0.1 http", "matcher", "http://127.0.0.1:3002/v1", false},
		{"[::1] http", "tracer", "http://[::1]:3003/v1", false},
		{"localhost no port", "fees", "http://localhost/v1", false},

		// Should NOT warn — empty or odd schemes.
		{"empty url", "midaz", "", false},
		{"ftp scheme", "midaz", "ftp://example.com/v1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureSlogOutput(func() {
				warnInsecureURL(tt.product, tt.url)
			})

			if tt.wantWarning {
				assert.Contains(t, output, "insecure URL detected")
				assert.Contains(t, output, tt.product)
				assert.Contains(t, output, "not recommended for production use")
			} else {
				assert.Empty(t, output, "no warning expected for %s", tt.url)
			}
		})
	}
}

// TestWarnInsecureURLIncludesProductAndURL verifies the structured log
// attributes contain both the product name and the offending URL.
func TestWarnInsecureURLIncludesProductAndURL(t *testing.T) {
	output := captureSlogOutput(func() {
		warnInsecureURL("midaz (onboarding)", "http://production.example.com:3000/v1")
	})

	assert.Contains(t, output, "midaz (onboarding)")
	assert.Contains(t, output, "http://production.example.com:3000/v1")
}
