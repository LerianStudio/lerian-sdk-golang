package lerian

import (
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/reporter"
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

	client, err := New(
		WithMidaz(), // signal intent; values come from env
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Midaz)
}

func TestMidazOAuth2EnvVarFallbacks(t *testing.T) {
	t.Setenv(envMidazOnboardingURL, "http://onboarding-from-env:3000/v1")
	t.Setenv(envMidazTransactionURL, "http://tx-from-env:3001/v1")
	t.Setenv(envMidazClientID, "env-client-id")
	t.Setenv(envMidazClientSecret, "env-client-secret")
	t.Setenv(envMidazTokenURL, "http://localhost:8080/token")

	client, err := New(
		WithMidaz(),
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, client.Midaz)

	assert.Equal(t, "env-client-id", midazConfigStringField(t, client, "ClientID"))
	assert.Equal(t, "env-client-secret", midazConfigStringField(t, client, "ClientSecret"))
	assert.Equal(t, "http://localhost:8080/token", midazConfigStringField(t, client, "TokenURL"))
}

func TestMidazExplicitOAuth2PartialConfigReturnsError(t *testing.T) {
	t.Setenv(envMidazOnboardingURL, "http://onboarding-from-env:3000/v1")
	t.Setenv(envMidazTransactionURL, "http://tx-from-env:3001/v1")
	t.Setenv(envMidazClientSecret, "env-client-secret")
	t.Setenv(envMidazTokenURL, "http://localhost:8080/token")

	_, err := New(
		WithMidaz(
			midaz.WithClientCredentials("explicit-client-id", "", ""),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
}

func TestMidazPartialOAuth2EnvConfigReturnsError(t *testing.T) {
	t.Setenv(envMidazOnboardingURL, "http://onboarding-from-env:3000/v1")
	t.Setenv(envMidazTransactionURL, "http://tx-from-env:3001/v1")
	t.Setenv(envMidazClientID, "env-client-id")
	t.Setenv(envMidazTokenURL, "http://localhost:8080/token")

	_, err := New(
		WithMidaz(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
}

func TestExplicitOverridesEnv(t *testing.T) {
	// Set env var values for Midaz.
	t.Setenv("LERIAN_MIDAZ_ONBOARDING_URL", "http://env-onboarding:3000/v1")
	t.Setenv("LERIAN_MIDAZ_TRANSACTION_URL", "http://env-transaction:3001/v1")
	t.Setenv("LERIAN_MIDAZ_CLIENT_ID", "env-client-id")
	t.Setenv("LERIAN_MIDAZ_CLIENT_SECRET", "env-client-secret")
	t.Setenv("LERIAN_MIDAZ_TOKEN_URL", "http://localhost:8080/token")

	// Provide explicit options — these should win over env vars.
	client, err := New(
		WithMidaz(
			midaz.WithOnboardingURL("http://explicit-onboarding:3000/v1"),
			midaz.WithTransactionURL("http://explicit-transaction:3001/v1"),
			midaz.WithClientCredentials("explicit-client-id", "explicit-client-secret", "http://localhost:9090/token"),
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

	// Client credentials from explicit option, URL from env.
	client, err := New(
		WithMatcher(
			matcher.WithClientCredentials("explicit-client-id", "explicit-client-secret", "http://localhost:8080/token"),
		),
	)
	require.NoError(t, err)
	assert.NotNil(t, client.Matcher)
}

func TestMatcherOAuth2EnvVars(t *testing.T) {
	t.Setenv(envMatcherURL, "http://matcher-from-env:3002/v1")
	t.Setenv(envMatcherClientID, "matcher-client-id")
	t.Setenv(envMatcherClientSecret, "matcher-client-secret")
	t.Setenv(envMatcherTokenURL, "http://localhost:8080/token")

	client, err := New(WithMatcher())
	require.NoError(t, err)
	require.NotNil(t, client.Matcher)
}

func TestMatcherPartialOAuth2EnvConfigReturnsError(t *testing.T) {
	t.Setenv(envMatcherURL, "http://matcher-from-env:3002/v1")
	t.Setenv(envMatcherClientID, "matcher-client-id")
	t.Setenv(envMatcherTokenURL, "http://localhost:8080/token")

	_, err := New(WithMatcher())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matcher: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
}

func TestMatcherExplicitOAuth2PartialConfigReturnsError(t *testing.T) {
	t.Setenv(envMatcherURL, "http://matcher-from-env:3002/v1")
	t.Setenv(envMatcherClientSecret, "matcher-client-secret")
	t.Setenv(envMatcherTokenURL, "http://localhost:8080/token")

	_, err := New(
		WithMatcher(
			matcher.WithClientCredentials("explicit-client-id", "", ""),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matcher: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
}

// ---------------------------------------------------------------------------
// Tracer env vars
// ---------------------------------------------------------------------------

func TestTracerEnvVars(t *testing.T) {
	t.Setenv("LERIAN_TRACER_URL", "http://tracer-from-env:3003/v1")

	client, err := New(
		WithTracer(), // signal intent; values come from env
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Tracer,
		"Tracer should be initialized from env vars")
}

func TestTracerEnvVarsMixedWithOptions(t *testing.T) {
	// URL from explicit option, client credentials from option.
	client, err := New(
		WithTracer(
			tracer.WithBaseURL("http://explicit-tracer:3003/v1"),
			tracer.WithClientCredentials("explicit-client-id", "explicit-client-secret", "http://localhost:8080/token"),
		),
	)
	require.NoError(t, err)
	assert.NotNil(t, client.Tracer)
}

func TestTracerOAuth2EnvVars(t *testing.T) {
	t.Setenv(envTracerURL, "http://tracer-from-env:3003/v1")
	t.Setenv(envTracerClientID, "tracer-client-id")
	t.Setenv(envTracerClientSecret, "tracer-client-secret")
	t.Setenv(envTracerTokenURL, "http://localhost:8080/token")

	client, err := New(WithTracer())
	require.NoError(t, err)
	require.NotNil(t, client.Tracer)
}

func TestTracerPartialOAuth2EnvConfigReturnsError(t *testing.T) {
	t.Setenv(envTracerURL, "http://tracer-from-env:3003/v1")
	t.Setenv(envTracerClientID, "tracer-client-id")
	t.Setenv(envTracerTokenURL, "http://localhost:8080/token")

	_, err := New(WithTracer())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tracer: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
}

func TestTracerExplicitOAuth2PartialConfigReturnsError(t *testing.T) {
	t.Setenv(envTracerURL, "http://tracer-from-env:3003/v1")
	t.Setenv(envTracerClientSecret, "tracer-client-secret")
	t.Setenv(envTracerTokenURL, "http://localhost:8080/token")

	_, err := New(
		WithTracer(
			tracer.WithClientCredentials("explicit-client-id", "", ""),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tracer: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
}

// ---------------------------------------------------------------------------
// Reporter env vars
// ---------------------------------------------------------------------------

func TestReporterEnvVars(t *testing.T) {
	t.Setenv("LERIAN_REPORTER_URL", "http://reporter-from-env:3004/v1")
	t.Setenv("LERIAN_REPORTER_ORG_ID", "org-from-env")

	client, err := New(
		WithReporter(), // signal intent; values come from env
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Reporter,
		"Reporter should be initialized from env vars")
}

func TestReporterOAuth2EnvVars(t *testing.T) {
	t.Setenv(envReporterURL, "http://reporter-from-env:3004/v1")
	t.Setenv(envReporterOrgID, "org-from-env")
	t.Setenv(envReporterClientID, "reporter-client-id")
	t.Setenv(envReporterClientSecret, "reporter-client-secret")
	t.Setenv(envReporterTokenURL, "http://localhost:8080/token")

	client, err := New(WithReporter())
	require.NoError(t, err)
	require.NotNil(t, client.Reporter)
}

func TestReporterPartialOAuth2EnvConfigReturnsError(t *testing.T) {
	t.Setenv(envReporterURL, "http://reporter-from-env:3004/v1")
	t.Setenv(envReporterOrgID, "org-from-env")
	t.Setenv(envReporterClientID, "reporter-client-id")
	t.Setenv(envReporterTokenURL, "http://localhost:8080/token")

	_, err := New(WithReporter())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reporter: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
}

func TestReporterExplicitOAuth2PartialConfigReturnsError(t *testing.T) {
	t.Setenv(envReporterURL, "http://reporter-from-env:3004/v1")
	t.Setenv(envReporterOrgID, "org-from-env")
	t.Setenv(envReporterClientSecret, "reporter-client-secret")
	t.Setenv(envReporterTokenURL, "http://localhost:8080/token")

	_, err := New(
		WithReporter(
			reporter.WithClientCredentials("explicit-client-id", "", ""),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reporter: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
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
	t.Setenv("LERIAN_FEES_ORG_ID", "org-fees-from-env")

	client, err := New(
		WithFees(), // signal intent; values come from env
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Fees,
		"Fees should be initialized from env vars")
}

func TestFeesOAuth2EnvVars(t *testing.T) {
	t.Setenv(envFeesURL, "http://fees-from-env:3005/v1")
	t.Setenv(envFeesOrgID, "org-fees-from-env")
	t.Setenv(envFeesClientID, "fees-client-id")
	t.Setenv(envFeesClientSecret, "fees-client-secret")
	t.Setenv(envFeesTokenURL, "http://localhost:8080/token")

	client, err := New(WithFees())
	require.NoError(t, err)
	require.NotNil(t, client.Fees)
}

func TestFeesPartialOAuth2EnvConfigReturnsError(t *testing.T) {
	t.Setenv(envFeesURL, "http://fees-from-env:3005/v1")
	t.Setenv(envFeesOrgID, "org-fees-from-env")
	t.Setenv(envFeesClientID, "fees-client-id")
	t.Setenv(envFeesTokenURL, "http://localhost:8080/token")

	_, err := New(WithFees())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fees: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
}

func TestFeesExplicitOAuth2PartialConfigReturnsError(t *testing.T) {
	t.Setenv(envFeesURL, "http://fees-from-env:3005/v1")
	t.Setenv(envFeesOrgID, "org-fees-from-env")
	t.Setenv(envFeesClientSecret, "fees-client-secret")
	t.Setenv(envFeesTokenURL, "http://localhost:8080/token")

	_, err := New(
		WithFees(
			fees.WithClientCredentials("explicit-client-id", "", ""),
		),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fees: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
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
	t.Setenv("LERIAN_FEES_CLIENT_ID", "env-client-id")
	t.Setenv("LERIAN_FEES_CLIENT_SECRET", "env-client-secret")
	t.Setenv("LERIAN_FEES_TOKEN_URL", "http://localhost:8081/token")
	t.Setenv("LERIAN_FEES_ORG_ID", "env-org")

	// Explicit options should win over env.
	client, err := New(
		WithFees(
			fees.WithBaseURL("http://explicit-fees:3005/v1"),
			fees.WithClientCredentials("explicit-client-id", "explicit-client-secret", "http://localhost:8080/token"),
			fees.WithOrganizationID("explicit-org"),
		),
	)
	require.NoError(t, err)
	assert.NotNil(t, client.Fees)
}

func TestExplicitProductConfigWithEnvOAuth2(t *testing.T) {
	t.Run("midaz", func(t *testing.T) {
		t.Setenv(envMidazClientID, "midaz-client-id")
		t.Setenv(envMidazClientSecret, "midaz-client-secret")
		t.Setenv(envMidazTokenURL, "http://localhost:8080/token")

		client, err := New(
			WithMidaz(
				midaz.WithOnboardingURL("http://explicit-onboarding:3000/v1"),
				midaz.WithTransactionURL("http://explicit-transaction:3001/v1"),
			),
		)
		require.NoError(t, err)
		require.NotNil(t, client.Midaz)
		assert.Equal(t, "midaz-client-id", midazConfigStringField(t, client, "ClientID"))
	})

	t.Run("matcher", func(t *testing.T) {
		t.Setenv(envMatcherClientID, "matcher-client-id")
		t.Setenv(envMatcherClientSecret, "matcher-client-secret")
		t.Setenv(envMatcherTokenURL, "http://localhost:8080/token")

		client, err := New(
			WithMatcher(
				matcher.WithBaseURL("http://explicit-matcher:3002/v1"),
			),
		)
		require.NoError(t, err)
		require.NotNil(t, client.Matcher)
		assert.Equal(t, "matcher-client-id", clientProductConfigStringField(t, client, "Matcher", "ClientID"))
	})

	t.Run("tracer", func(t *testing.T) {
		t.Setenv(envTracerClientID, "tracer-client-id")
		t.Setenv(envTracerClientSecret, "tracer-client-secret")
		t.Setenv(envTracerTokenURL, "http://localhost:8080/token")

		client, err := New(
			WithTracer(
				tracer.WithBaseURL("http://explicit-tracer:3003/v1"),
			),
		)
		require.NoError(t, err)
		require.NotNil(t, client.Tracer)
		assert.Equal(t, "tracer-client-id", clientProductConfigStringField(t, client, "Tracer", "ClientID"))
	})

	t.Run("reporter", func(t *testing.T) {
		t.Setenv(envReporterClientID, "reporter-client-id")
		t.Setenv(envReporterClientSecret, "reporter-client-secret")
		t.Setenv(envReporterTokenURL, "http://localhost:8080/token")

		client, err := New(
			WithReporter(
				reporter.WithBaseURL("http://explicit-reporter:3004/v1"),
				reporter.WithOrganizationID("org-explicit"),
			),
		)
		require.NoError(t, err)
		require.NotNil(t, client.Reporter)
		assert.Equal(t, "reporter-client-id", clientProductConfigStringField(t, client, "Reporter", "ClientID"))
	})

	t.Run("fees", func(t *testing.T) {
		t.Setenv(envFeesClientID, "fees-client-id")
		t.Setenv(envFeesClientSecret, "fees-client-secret")
		t.Setenv(envFeesTokenURL, "http://localhost:8080/token")

		client, err := New(
			WithFees(
				fees.WithBaseURL("http://explicit-fees:3005/v1"),
				fees.WithOrganizationID("org-explicit"),
			),
		)
		require.NoError(t, err)
		require.NotNil(t, client.Fees)
		assert.Equal(t, "fees-client-id", clientProductConfigStringField(t, client, "Fees", "ClientID"))
	})
}

func TestExplicitProductConfigWithPartialEnvOAuth2ReturnsError(t *testing.T) {
	t.Run("midaz", func(t *testing.T) {
		t.Setenv(envMidazClientID, "midaz-client-id")
		t.Setenv(envMidazTokenURL, "http://localhost:8080/token")

		_, err := New(
			WithMidaz(
				midaz.WithOnboardingURL("http://explicit-onboarding:3000/v1"),
				midaz.WithTransactionURL("http://explicit-transaction:3001/v1"),
			),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
	})

	t.Run("matcher", func(t *testing.T) {
		t.Setenv(envMatcherClientID, "matcher-client-id")
		t.Setenv(envMatcherTokenURL, "http://localhost:8080/token")

		_, err := New(
			WithMatcher(
				matcher.WithBaseURL("http://explicit-matcher:3002/v1"),
			),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "matcher: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
	})

	t.Run("tracer", func(t *testing.T) {
		t.Setenv(envTracerClientID, "tracer-client-id")
		t.Setenv(envTracerTokenURL, "http://localhost:8080/token")

		_, err := New(
			WithTracer(
				tracer.WithBaseURL("http://explicit-tracer:3003/v1"),
			),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tracer: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
	})

	t.Run("reporter", func(t *testing.T) {
		t.Setenv(envReporterClientID, "reporter-client-id")
		t.Setenv(envReporterTokenURL, "http://localhost:8080/token")

		_, err := New(
			WithReporter(
				reporter.WithBaseURL("http://explicit-reporter:3004/v1"),
				reporter.WithOrganizationID("org-explicit"),
			),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reporter: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
	})

	t.Run("fees", func(t *testing.T) {
		t.Setenv(envFeesClientID, "fees-client-id")
		t.Setenv(envFeesTokenURL, "http://localhost:8080/token")

		_, err := New(
			WithFees(
				fees.WithBaseURL("http://explicit-fees:3005/v1"),
				fees.WithOrganizationID("org-explicit"),
			),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fees: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
	})
}

func TestLegacyLerianAuthEnvVarsReturnMigrationError(t *testing.T) {
	t.Run("midaz", func(t *testing.T) {
		t.Setenv(envMidazLegacyAuthToken, "legacy-token")

		_, err := New(
			WithMidaz(
				midaz.WithOnboardingURL("http://explicit-onboarding:3000/v1"),
				midaz.WithTransactionURL("http://explicit-transaction:3001/v1"),
			),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), envMidazLegacyAuthToken)
	})

	t.Run("matcher", func(t *testing.T) {
		t.Setenv(envMatcherLegacyAPIKey, "legacy-key")

		_, err := New(
			WithMatcher(
				matcher.WithBaseURL("http://explicit-matcher:3002/v1"),
			),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), envMatcherLegacyAPIKey)
	})

	t.Run("tracer", func(t *testing.T) {
		t.Setenv(envTracerLegacyAPIKey, "legacy-key")

		_, err := New(
			WithTracer(
				tracer.WithBaseURL("http://explicit-tracer:3003/v1"),
			),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), envTracerLegacyAPIKey)
	})

	t.Run("reporter", func(t *testing.T) {
		t.Setenv(envReporterLegacyAuthToken, "legacy-token")

		_, err := New(
			WithReporter(
				reporter.WithBaseURL("http://explicit-reporter:3004/v1"),
				reporter.WithOrganizationID("org-explicit"),
			),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), envReporterLegacyAuthToken)
	})

	t.Run("fees", func(t *testing.T) {
		t.Setenv(envFeesLegacyAuthToken, "legacy-token")

		_, err := New(
			WithFees(
				fees.WithBaseURL("http://explicit-fees:3005/v1"),
				fees.WithOrganizationID("org-explicit"),
			),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), envFeesLegacyAuthToken)
	})
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
	t.Setenv("LERIAN_MIDAZ_CLIENT_ID", "midaz-client-id")
	t.Setenv("LERIAN_MIDAZ_CLIENT_SECRET", "midaz-client-secret")
	t.Setenv("LERIAN_MIDAZ_TOKEN_URL", "http://localhost:8080/token")

	t.Setenv("LERIAN_MATCHER_URL", "http://matcher:3002/v1")
	t.Setenv("LERIAN_MATCHER_CLIENT_ID", "matcher-client-id")
	t.Setenv("LERIAN_MATCHER_CLIENT_SECRET", "matcher-client-secret")
	t.Setenv("LERIAN_MATCHER_TOKEN_URL", "http://localhost:8080/token")

	t.Setenv("LERIAN_TRACER_URL", "http://tracer:3003/v1")
	t.Setenv("LERIAN_TRACER_CLIENT_ID", "tracer-client-id")
	t.Setenv("LERIAN_TRACER_CLIENT_SECRET", "tracer-client-secret")
	t.Setenv("LERIAN_TRACER_TOKEN_URL", "http://localhost:8080/token")

	t.Setenv("LERIAN_REPORTER_URL", "http://reporter:3004/v1")
	t.Setenv("LERIAN_REPORTER_ORG_ID", "org-reporter")
	t.Setenv("LERIAN_REPORTER_CLIENT_ID", "reporter-client-id")
	t.Setenv("LERIAN_REPORTER_CLIENT_SECRET", "reporter-client-secret")
	t.Setenv("LERIAN_REPORTER_TOKEN_URL", "http://localhost:8080/token")

	t.Setenv("LERIAN_FEES_URL", "http://fees:3005/v1")
	t.Setenv("LERIAN_FEES_ORG_ID", "org-fees")
	t.Setenv("LERIAN_FEES_CLIENT_ID", "fees-client-id")
	t.Setenv("LERIAN_FEES_CLIENT_SECRET", "fees-client-secret")
	t.Setenv("LERIAN_FEES_TOKEN_URL", "http://localhost:8080/token")

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
	t.Parallel()

	// Verify the constant values match the documented LERIAN_* prefix.
	// This guards against accidental renames.
	assert.Equal(t, "LERIAN_DEBUG", envDebug)

	assert.Equal(t, "LERIAN_MIDAZ_ONBOARDING_URL", envMidazOnboardingURL)
	assert.Equal(t, "LERIAN_MIDAZ_TRANSACTION_URL", envMidazTransactionURL)
	assert.Equal(t, "LERIAN_MIDAZ_CLIENT_ID", envMidazClientID)
	assert.Equal(t, "LERIAN_MIDAZ_CLIENT_SECRET", envMidazClientSecret)
	assert.Equal(t, "LERIAN_MIDAZ_TOKEN_URL", envMidazTokenURL)

	assert.Equal(t, "LERIAN_MATCHER_URL", envMatcherURL)
	assert.Equal(t, "LERIAN_MATCHER_CLIENT_ID", envMatcherClientID)
	assert.Equal(t, "LERIAN_MATCHER_CLIENT_SECRET", envMatcherClientSecret)
	assert.Equal(t, "LERIAN_MATCHER_TOKEN_URL", envMatcherTokenURL)

	assert.Equal(t, "LERIAN_TRACER_URL", envTracerURL)
	assert.Equal(t, "LERIAN_TRACER_CLIENT_ID", envTracerClientID)
	assert.Equal(t, "LERIAN_TRACER_CLIENT_SECRET", envTracerClientSecret)
	assert.Equal(t, "LERIAN_TRACER_TOKEN_URL", envTracerTokenURL)

	assert.Equal(t, "LERIAN_REPORTER_URL", envReporterURL)
	assert.Equal(t, "LERIAN_REPORTER_CLIENT_ID", envReporterClientID)
	assert.Equal(t, "LERIAN_REPORTER_CLIENT_SECRET", envReporterClientSecret)
	assert.Equal(t, "LERIAN_REPORTER_TOKEN_URL", envReporterTokenURL)
	assert.Equal(t, "LERIAN_REPORTER_ORG_ID", envReporterOrgID)

	assert.Equal(t, "LERIAN_FEES_URL", envFeesURL)
	assert.Equal(t, "LERIAN_FEES_CLIENT_ID", envFeesClientID)
	assert.Equal(t, "LERIAN_FEES_CLIENT_SECRET", envFeesClientSecret)
	assert.Equal(t, "LERIAN_FEES_TOKEN_URL", envFeesTokenURL)
	assert.Equal(t, "LERIAN_FEES_ORG_ID", envFeesOrgID)
}
