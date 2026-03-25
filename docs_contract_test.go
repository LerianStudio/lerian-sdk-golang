package lerian

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDotEnvExampleMatchesSupportedOAuthEnvVars(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(".env.example")
	require.NoError(t, err)

	text := string(content)
	for _, key := range []string{
		envDebug,
		envMidazOnboardingURL,
		envMidazTransactionURL,
		envMidazClientID,
		envMidazClientSecret,
		envMidazTokenURL,
		envMatcherURL,
		envMatcherClientID,
		envMatcherClientSecret,
		envMatcherTokenURL,
		envTracerURL,
		envTracerClientID,
		envTracerClientSecret,
		envTracerTokenURL,
		envReporterURL,
		envReporterOrgID,
		envReporterClientID,
		envReporterClientSecret,
		envReporterTokenURL,
		envFeesURL,
		envFeesOrgID,
		envFeesClientID,
		envFeesClientSecret,
		envFeesTokenURL,
	} {
		assert.Contains(t, text, key)
	}

	for _, legacyKey := range []string{
		"LERIAN_MIDAZ_AUTH_TOKEN",
		"LERIAN_MATCHER_API_KEY",
		"LERIAN_TRACER_API_KEY",
		"LERIAN_REPORTER_AUTH_TOKEN",
		"LERIAN_FEES_AUTH_TOKEN",
	} {
		assert.NotContains(t, text, legacyKey)
	}
}

func TestREADMEOAuthExamplesMatchCurrentAPI(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("README.md")
	require.NoError(t, err)

	text := string(content)
	for _, expected := range []string{
		"defer client.Shutdown(ctx)",
		"&midaz.CreateOrganizationInput{",
		"OnboardingURL:  \"http://localhost:3000/v1\"",
		"TransactionURL: \"http://localhost:3001/v1\"",
		"Matcher: &matcher.Config{BaseURL: \"http://localhost:3002/v1\"}",
		"Tracer:  &tracer.Config{BaseURL: \"http://localhost:3003/v1\"}",
		"BaseURL:        \"http://localhost:3004/v1\"",
		"BaseURL:        \"http://localhost:3005/v1\"",
		"retryCfg.MaxRetries = 3",
		"ObservabilityConfig{",
		"LoadConfigFromEnv()",
		"LERIAN_*_CLIENT_ID",
	} {
		assert.Contains(t, text, expected)
	}

	for _, unexpected := range []string{
		"defer client.Close(ctx)",
		"models.CreateOrganizationInput{",
		"WithRetryConfig(",
		"WithObservability(observability.Config{",
		"lerian.WithMidaz(",
	} {
		assert.NotContains(t, text, unexpected)
	}
}

func TestClaudeAuthGuidanceMatchesCurrentEnvContract(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("CLAUDE.md")
	require.NoError(t, err)

	text := string(content)
	assert.Contains(t, text, envMidazClientID)
	assert.NotContains(t, text, "LERIAN_MIDAZ_AUTH_TOKEN")
}
