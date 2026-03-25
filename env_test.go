package lerian

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvOrDefault(t *testing.T) {
	t.Run("returns env value when set", func(t *testing.T) {
		t.Setenv("LERIAN_TEST_ENVORD_SET", "from-env")
		assert.Equal(t, "from-env", envOrDefault("LERIAN_TEST_ENVORD_SET", "fallback"))
	})

	t.Run("returns default when empty or unset", func(t *testing.T) {
		t.Setenv("LERIAN_TEST_ENVORD_EMPTY", "")
		assert.Equal(t, "fallback", envOrDefault("LERIAN_TEST_ENVORD_EMPTY", "fallback"))
		assert.Equal(t, "fallback", envOrDefault("LERIAN_TEST_ENVORD_UNSET", "fallback"))
	})
}

func TestEnvBool(t *testing.T) {
	t.Setenv("LERIAN_TEST_BOOL_TRUE", "true")
	t.Setenv("LERIAN_TEST_BOOL_ONE", "1")
	t.Setenv("LERIAN_TEST_BOOL_FALSE", "false")

	assert.True(t, envBool("LERIAN_TEST_BOOL_TRUE"))
	assert.True(t, envBool("LERIAN_TEST_BOOL_ONE"))
	assert.False(t, envBool("LERIAN_TEST_BOOL_FALSE"))
	assert.False(t, envBool("LERIAN_TEST_BOOL_UNSET"))
}

func TestLoadConfigFromEnvEmpty(t *testing.T) {
	t.Parallel()

	cfg := LoadConfigFromEnv()
	assert.Nil(t, cfg.Midaz)
	assert.Nil(t, cfg.Matcher)
	assert.Nil(t, cfg.Tracer)
	assert.Nil(t, cfg.Reporter)
	assert.Nil(t, cfg.Fees)
	assert.False(t, cfg.Debug)
}

func TestLoadConfigFromEnvDebug(t *testing.T) {
	t.Setenv(envDebug, "true")

	cfg := LoadConfigFromEnv()
	assert.True(t, cfg.Debug)
}

func TestLoadConfigFromEnvMidaz(t *testing.T) {
	t.Setenv(envMidazOnboardingURL, "http://onboarding-from-env:3000/v1")
	t.Setenv(envMidazTransactionURL, "http://tx-from-env:3001/v1")
	t.Setenv(envMidazCRMURL, "http://crm-from-env:4003/v1")
	t.Setenv(envMidazClientID, "env-client-id")
	t.Setenv(envMidazClientSecret, "env-client-secret")
	t.Setenv(envMidazTokenURL, "http://localhost:8080/token")

	cfg := LoadConfigFromEnv()
	require.NotNil(t, cfg.Midaz)
	assert.Equal(t, "http://onboarding-from-env:3000/v1", cfg.Midaz.OnboardingURL)
	assert.Equal(t, "http://tx-from-env:3001/v1", cfg.Midaz.TransactionURL)
	assert.Equal(t, "http://crm-from-env:4003/v1", cfg.Midaz.CRMURL)
	assert.Equal(t, "env-client-id", cfg.Midaz.ClientID)
}

func TestLoadConfigFromEnvOtherProducts(t *testing.T) {
	t.Setenv(envMatcherURL, "http://matcher-from-env:3002/v1")
	t.Setenv(envTracerURL, "http://tracer-from-env:3003/v1")
	t.Setenv(envReporterURL, "http://reporter-from-env:3004/v1")
	t.Setenv(envReporterOrgID, "org-reporter")
	t.Setenv(envFeesURL, "http://fees-from-env:3005/v1")
	t.Setenv(envFeesOrgID, "org-fees")

	cfg := LoadConfigFromEnv()
	require.NotNil(t, cfg.Matcher)
	require.NotNil(t, cfg.Tracer)
	require.NotNil(t, cfg.Reporter)
	require.NotNil(t, cfg.Fees)
	assert.Equal(t, "http://matcher-from-env:3002/v1", cfg.Matcher.BaseURL)
	assert.Equal(t, "http://tracer-from-env:3003/v1", cfg.Tracer.BaseURL)
	assert.Equal(t, "org-reporter", cfg.Reporter.OrganizationID)
	assert.Equal(t, "org-fees", cfg.Fees.OrganizationID)
}

func TestNewFromLoadedEnvConfig(t *testing.T) {
	t.Setenv(envMidazOnboardingURL, "http://midaz-onboarding:3000/v1")
	t.Setenv(envMidazTransactionURL, "http://midaz-transaction:3001/v1")
	t.Setenv(envMatcherURL, "http://matcher:3002/v1")
	t.Setenv(envTracerURL, "http://tracer:3003/v1")
	t.Setenv(envReporterURL, "http://reporter:3004/v1")
	t.Setenv(envReporterOrgID, "org-reporter")
	t.Setenv(envFeesURL, "http://fees:3005/v1")
	t.Setenv(envFeesOrgID, "org-fees")
	t.Setenv(envDebug, "1")

	client, err := New(LoadConfigFromEnv())
	require.NoError(t, err)
	require.NotNil(t, client.Midaz)
	require.NotNil(t, client.Matcher)
	require.NotNil(t, client.Tracer)
	require.NotNil(t, client.Reporter)
	require.NotNil(t, client.Fees)
	assert.True(t, client.debug)
}

func TestLoadConfigFromEnvTriggersValidationForPartialProductConfig(t *testing.T) {
	t.Setenv(envMatcherClientID, "matcher-client-id")
	t.Setenv(envMatcherTokenURL, "http://localhost:8080/token")

	_, err := New(LoadConfigFromEnv())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matcher: ClientID, ClientSecret, and TokenURL must all be set for OAuth2")
}

func TestLoadConfigFromEnvTriggersLegacyMigrationErrors(t *testing.T) {
	t.Setenv(envMidazLegacyAuthToken, "legacy-token")

	_, err := New(LoadConfigFromEnv())
	require.Error(t, err)
	assert.Contains(t, err.Error(), envMidazLegacyAuthToken)
}

func TestNoMidazEnvVars(t *testing.T) {
	t.Setenv("MIDAZ_ONBOARDING_URL", "http://v2-onboarding:3000/v1")
	t.Setenv("MIDAZ_TRANSACTION_URL", "http://v2-transaction:3001/v1")
	t.Setenv("MIDAZ_AUTH_TOKEN", "v2-token")

	cfg := LoadConfigFromEnv()
	assert.Nil(t, cfg.Midaz)
}

func TestEnvVarConstantValues(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "LERIAN_DEBUG", envDebug)
	assert.Equal(t, "LERIAN_MIDAZ_ONBOARDING_URL", envMidazOnboardingURL)
	assert.Equal(t, "LERIAN_MIDAZ_TRANSACTION_URL", envMidazTransactionURL)
	assert.Equal(t, "LERIAN_MIDAZ_CRM_URL", envMidazCRMURL)
	assert.Equal(t, "LERIAN_MIDAZ_CLIENT_ID", envMidazClientID)
	assert.Equal(t, "LERIAN_MIDAZ_CLIENT_SECRET", envMidazClientSecret)
	assert.Equal(t, "LERIAN_MIDAZ_TOKEN_URL", envMidazTokenURL)
	assert.Equal(t, "LERIAN_MATCHER_URL", envMatcherURL)
	assert.Equal(t, "LERIAN_TRACER_URL", envTracerURL)
	assert.Equal(t, "LERIAN_REPORTER_URL", envReporterURL)
	assert.Equal(t, "LERIAN_FEES_URL", envFeesURL)
}
