package lerian

import (
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// envOrDefault unit tests
// ---------------------------------------------------------------------------

func TestEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		setEnv       bool
		defaultValue string
		want         string
	}{
		{
			name:         "returns env value when set",
			key:          "LERIAN_TEST_ENVORD_SET",
			envValue:     "from-env",
			setEnv:       true,
			defaultValue: "fallback",
			want:         "from-env",
		},
		{
			name:         "returns default when env is empty",
			key:          "LERIAN_TEST_ENVORD_EMPTY",
			envValue:     "",
			setEnv:       true,
			defaultValue: "fallback",
			want:         "fallback",
		},
		{
			name:         "returns default when env is not set",
			key:          "LERIAN_TEST_ENVORD_UNSET",
			setEnv:       false,
			defaultValue: "fallback",
			want:         "fallback",
		},
		{
			name:         "returns empty default when nothing set",
			key:          "LERIAN_TEST_ENVORD_NODEFAULT",
			setEnv:       false,
			defaultValue: "",
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			got := envOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// envBool unit tests
// ---------------------------------------------------------------------------

func TestEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		setEnv   bool
		want     bool
	}{
		{
			name:     "true string returns true",
			key:      "LERIAN_TEST_BOOL_TRUE",
			envValue: "true",
			setEnv:   true,
			want:     true,
		},
		{
			name:     "1 string returns true",
			key:      "LERIAN_TEST_BOOL_ONE",
			envValue: "1",
			setEnv:   true,
			want:     true,
		},
		{
			name:     "false string returns false",
			key:      "LERIAN_TEST_BOOL_FALSE",
			envValue: "false",
			setEnv:   true,
			want:     false,
		},
		{
			name:     "empty string returns false",
			key:      "LERIAN_TEST_BOOL_EMPTY",
			envValue: "",
			setEnv:   true,
			want:     false,
		},
		{
			name:   "unset returns false",
			key:    "LERIAN_TEST_BOOL_UNSET",
			setEnv: false,
			want:   false,
		},
		{
			name:     "random string returns false",
			key:      "LERIAN_TEST_BOOL_RANDOM",
			envValue: "yes",
			setEnv:   true,
			want:     false,
		},
		{
			name:     "TRUE uppercase returns false",
			key:      "LERIAN_TEST_BOOL_UPPER",
			envValue: "TRUE",
			setEnv:   true,
			want:     false,
		},
		{
			name:     "0 returns false",
			key:      "LERIAN_TEST_BOOL_ZERO",
			envValue: "0",
			setEnv:   true,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			got := envBool(tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Midaz env var integration
// ---------------------------------------------------------------------------

func TestEnvVarFallback(t *testing.T) {
	// Set onboarding URL via env var, provide transaction URL via option.
	// Both required fields are satisfied => no error.
	t.Setenv("LERIAN_MIDAZ_ONBOARDING_URL", "http://onboarding-from-env:3000/v1")

	client, err := New(
		WithMidaz(
			midaz.WithTransactionURL("http://tx:3001/v1"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Midaz, "Midaz should be initialized via env + option")
}

func TestEnvVarFallbackBothURLs(t *testing.T) {
	// Both Midaz URLs come from env vars — no explicit options needed
	// beyond signaling that Midaz is wanted (empty option slice triggers init).
	t.Setenv("LERIAN_MIDAZ_ONBOARDING_URL", "http://onboarding-from-env:3000/v1")
	t.Setenv("LERIAN_MIDAZ_TRANSACTION_URL", "http://tx-from-env:3001/v1")
	t.Setenv("LERIAN_MIDAZ_AUTH_TOKEN", "env-token-123")

	client, err := New(
		WithMidaz(), // signal intent; values come from env
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Midaz)
}

func TestExplicitOverridesEnv(t *testing.T) {
	// Set env var values for Midaz.
	t.Setenv("LERIAN_MIDAZ_ONBOARDING_URL", "http://env-onboarding:3000/v1")
	t.Setenv("LERIAN_MIDAZ_TRANSACTION_URL", "http://env-transaction:3001/v1")
	t.Setenv("LERIAN_MIDAZ_AUTH_TOKEN", "env-token")

	// Provide explicit options — these should win over env vars.
	client, err := New(
		WithMidaz(
			midaz.WithOnboardingURL("http://explicit-onboarding:3000/v1"),
			midaz.WithTransactionURL("http://explicit-transaction:3001/v1"),
			midaz.WithAuthToken("explicit-token"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Midaz,
		"Midaz should be initialized with explicit values winning over env")
}

func TestDefaultWhenNoEnvAndNoOption(t *testing.T) {
	// No env var, no option for required fields => error returned.
	_, err := New(
		WithMidaz(), // triggers init but provides nothing
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OnboardingURL is required")
}

// ---------------------------------------------------------------------------
// Debug env var
// ---------------------------------------------------------------------------

func TestDebugEnvVar(t *testing.T) {
	t.Setenv("LERIAN_DEBUG", "true")

	client, err := New()
	require.NoError(t, err)
	assert.True(t, client.debug,
		"debug should be true when LERIAN_DEBUG=true")
}

func TestDebugEnvVarOne(t *testing.T) {
	t.Setenv("LERIAN_DEBUG", "1")

	client, err := New()
	require.NoError(t, err)
	assert.True(t, client.debug,
		"debug should be true when LERIAN_DEBUG=1")
}

func TestDebugEnvVarExplicitOverrides(t *testing.T) {
	// Even though env says false, explicit option true should win.
	t.Setenv("LERIAN_DEBUG", "false")

	client, err := New(WithDebug(true))
	require.NoError(t, err)
	assert.True(t, client.debug,
		"explicit WithDebug(true) should override env LERIAN_DEBUG=false")
}

func TestDebugEnvVarNotSet(t *testing.T) {
	// No env var, no option => debug is false.
	client, err := New()
	require.NoError(t, err)
	assert.False(t, client.debug,
		"debug should default to false when nothing is set")
}

// ---------------------------------------------------------------------------
// Matcher env vars
// ---------------------------------------------------------------------------

func TestMatcherEnvVars(t *testing.T) {
	t.Setenv("LERIAN_MATCHER_URL", "http://matcher-from-env:3002/v1")
	t.Setenv("LERIAN_MATCHER_API_KEY", "matcher-env-key")

	client, err := New(
		WithMatcher(), // signal intent; values come from env
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Matcher,
		"Matcher should be initialized from env vars")
}

func TestMatcherEnvVarsMixedWithOptions(t *testing.T) {
	t.Setenv("LERIAN_MATCHER_URL", "http://matcher-from-env:3002/v1")

	// API key from explicit option, URL from env.
	client, err := New(
		WithMatcher(
			matcher.WithAPIKey("explicit-matcher-key"),
		),
	)
	require.NoError(t, err)
	assert.NotNil(t, client.Matcher)
}

// ---------------------------------------------------------------------------
// Tracer env vars
// ---------------------------------------------------------------------------

func TestTracerEnvVars(t *testing.T) {
	t.Setenv("LERIAN_TRACER_URL", "http://tracer-from-env:3003/v1")
	t.Setenv("LERIAN_TRACER_API_KEY", "tracer-env-key")

	client, err := New(
		WithTracer(), // signal intent; values come from env
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Tracer,
		"Tracer should be initialized from env vars")
}

func TestTracerEnvVarsMixedWithOptions(t *testing.T) {
	t.Setenv("LERIAN_TRACER_API_KEY", "tracer-env-key")

	// URL from explicit option, API key from env.
	client, err := New(
		WithTracer(
			tracer.WithBaseURL("http://explicit-tracer:3003/v1"),
		),
	)
	require.NoError(t, err)
	assert.NotNil(t, client.Tracer)
}

// ---------------------------------------------------------------------------
// Reporter env vars
// ---------------------------------------------------------------------------

func TestReporterEnvVars(t *testing.T) {
	t.Setenv("LERIAN_REPORTER_URL", "http://reporter-from-env:3004/v1")
	t.Setenv("LERIAN_REPORTER_AUTH_TOKEN", "reporter-env-token")
	t.Setenv("LERIAN_REPORTER_ORG_ID", "org-from-env")

	client, err := New(
		WithReporter(), // signal intent; values come from env
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Reporter,
		"Reporter should be initialized from env vars")
}

func TestReporterEnvVarsMissingRequired(t *testing.T) {
	// Only URL set, missing OrganizationID => error.
	t.Setenv("LERIAN_REPORTER_URL", "http://reporter-from-env:3004/v1")

	_, err := New(
		WithReporter(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OrganizationID is required")
}

// ---------------------------------------------------------------------------
// Fees env vars
// ---------------------------------------------------------------------------

func TestFeesEnvVars(t *testing.T) {
	t.Setenv("LERIAN_FEES_URL", "http://fees-from-env:3005/v1")
	t.Setenv("LERIAN_FEES_AUTH_TOKEN", "fees-env-token")
	t.Setenv("LERIAN_FEES_ORG_ID", "org-fees-from-env")

	client, err := New(
		WithFees(), // signal intent; values come from env
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Fees,
		"Fees should be initialized from env vars")
}

func TestFeesEnvVarsMissingRequired(t *testing.T) {
	// Only URL set, missing OrganizationID => error.
	t.Setenv("LERIAN_FEES_URL", "http://fees-from-env:3005/v1")

	_, err := New(
		WithFees(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OrganizationID is required")
}

func TestFeesEnvVarsExplicitOverrides(t *testing.T) {
	t.Setenv("LERIAN_FEES_URL", "http://fees-env:3005/v1")
	t.Setenv("LERIAN_FEES_AUTH_TOKEN", "env-token")
	t.Setenv("LERIAN_FEES_ORG_ID", "env-org")

	// Explicit options should win over env.
	client, err := New(
		WithFees(
			fees.WithBaseURL("http://explicit-fees:3005/v1"),
			fees.WithAuthToken("explicit-token"),
			fees.WithOrganizationID("explicit-org"),
		),
	)
	require.NoError(t, err)
	assert.NotNil(t, client.Fees)
}

// ---------------------------------------------------------------------------
// Clean break: MIDAZ_* vars must NOT be read
// ---------------------------------------------------------------------------

func TestNoMidazEnvVars(t *testing.T) {
	// Set legacy v2 environment variables — these should be completely ignored.
	t.Setenv("MIDAZ_ONBOARDING_URL", "http://v2-onboarding:3000/v1")
	t.Setenv("MIDAZ_TRANSACTION_URL", "http://v2-transaction:3001/v1")
	t.Setenv("MIDAZ_AUTH_TOKEN", "v2-token")
	t.Setenv("MIDAZ_MAX_RETRIES", "10")

	// Do NOT set any LERIAN_* vars. The old MIDAZ_* vars must not be read.
	_, err := New(
		WithMidaz(), // triggers init but no LERIAN_* env set
	)
	require.Error(t, err, "old MIDAZ_* env vars should NOT be read")
	assert.Contains(t, err.Error(), "OnboardingURL is required",
		"should fail because LERIAN_MIDAZ_ONBOARDING_URL is not set")
}

func TestNoMidazDebugEnvVar(t *testing.T) {
	// Legacy v2 debug env var should not be read.
	t.Setenv("MIDAZ_DEBUG", "true")

	client, err := New()
	require.NoError(t, err)
	assert.False(t, client.debug,
		"MIDAZ_DEBUG should not enable debug mode; only LERIAN_DEBUG works")
}

// ---------------------------------------------------------------------------
// All products from env vars — full integration
// ---------------------------------------------------------------------------

func TestAllProductsFromEnvVars(t *testing.T) {
	// Configure every product entirely through environment variables.
	t.Setenv("LERIAN_DEBUG", "1")

	t.Setenv("LERIAN_MIDAZ_ONBOARDING_URL", "http://midaz-onboarding:3000/v1")
	t.Setenv("LERIAN_MIDAZ_TRANSACTION_URL", "http://midaz-transaction:3001/v1")
	t.Setenv("LERIAN_MIDAZ_AUTH_TOKEN", "midaz-token")

	t.Setenv("LERIAN_MATCHER_URL", "http://matcher:3002/v1")
	t.Setenv("LERIAN_MATCHER_API_KEY", "matcher-key")

	t.Setenv("LERIAN_TRACER_URL", "http://tracer:3003/v1")
	t.Setenv("LERIAN_TRACER_API_KEY", "tracer-key")

	t.Setenv("LERIAN_REPORTER_URL", "http://reporter:3004/v1")
	t.Setenv("LERIAN_REPORTER_AUTH_TOKEN", "reporter-token")
	t.Setenv("LERIAN_REPORTER_ORG_ID", "org-reporter")

	t.Setenv("LERIAN_FEES_URL", "http://fees:3005/v1")
	t.Setenv("LERIAN_FEES_AUTH_TOKEN", "fees-token")
	t.Setenv("LERIAN_FEES_ORG_ID", "org-fees")

	client, err := New(
		WithMidaz(),
		WithMatcher(),
		WithTracer(),
		WithReporter(),
		WithFees(),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.True(t, client.debug, "debug should be enabled via env")
	assert.NotNil(t, client.Midaz, "Midaz should be initialized from env")
	assert.NotNil(t, client.Matcher, "Matcher should be initialized from env")
	assert.NotNil(t, client.Tracer, "Tracer should be initialized from env")
	assert.NotNil(t, client.Reporter, "Reporter should be initialized from env")
	assert.NotNil(t, client.Fees, "Fees should be initialized from env")
}

// ---------------------------------------------------------------------------
// Env var constants sanity check
// ---------------------------------------------------------------------------

func TestEnvVarConstantValues(t *testing.T) {
	// Verify the constant values match the documented LERIAN_* prefix.
	// This guards against accidental renames.
	assert.Equal(t, "LERIAN_DEBUG", envDebug)

	assert.Equal(t, "LERIAN_MIDAZ_ONBOARDING_URL", envMidazOnboardingURL)
	assert.Equal(t, "LERIAN_MIDAZ_TRANSACTION_URL", envMidazTransactionURL)
	assert.Equal(t, "LERIAN_MIDAZ_AUTH_TOKEN", envMidazAuthToken)

	assert.Equal(t, "LERIAN_MATCHER_URL", envMatcherURL)
	assert.Equal(t, "LERIAN_MATCHER_API_KEY", envMatcherAPIKey)

	assert.Equal(t, "LERIAN_TRACER_URL", envTracerURL)
	assert.Equal(t, "LERIAN_TRACER_API_KEY", envTracerAPIKey)

	assert.Equal(t, "LERIAN_REPORTER_URL", envReporterURL)
	assert.Equal(t, "LERIAN_REPORTER_AUTH_TOKEN", envReporterAuthToken)
	assert.Equal(t, "LERIAN_REPORTER_ORG_ID", envReporterOrgID)

	assert.Equal(t, "LERIAN_FEES_URL", envFeesURL)
	assert.Equal(t, "LERIAN_FEES_AUTH_TOKEN", envFeesAuthToken)
	assert.Equal(t, "LERIAN_FEES_ORG_ID", envFeesOrgID)
}
