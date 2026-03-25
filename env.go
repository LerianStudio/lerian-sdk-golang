package lerian

import (
	"os"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/matcher"
	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/reporter"
	"github.com/LerianStudio/lerian-sdk-golang/tracer"
)

// ---------------------------------------------------------------------------
// Environment variable constants
// ---------------------------------------------------------------------------

// LERIAN_* environment variables are used by [LoadConfigFromEnv] to build a
// root [Config]. Only the LERIAN_* prefix is recognized; legacy MIDAZ_* vars
// from v2 are deliberately ignored to enforce a clean break.
const (
	// Debug
	envDebug = "LERIAN_DEBUG"

	// Midaz
	envMidazOnboardingURL   = "LERIAN_MIDAZ_ONBOARDING_URL"
	envMidazTransactionURL  = "LERIAN_MIDAZ_TRANSACTION_URL"
	envMidazCRMURL          = "LERIAN_MIDAZ_CRM_URL"
	envMidazClientID        = "LERIAN_MIDAZ_CLIENT_ID"
	envMidazClientSecret    = "LERIAN_MIDAZ_CLIENT_SECRET"
	envMidazTokenURL        = "LERIAN_MIDAZ_TOKEN_URL"
	envMidazLegacyAuthToken = "LERIAN_MIDAZ_AUTH_TOKEN"

	// Matcher
	envMatcherURL          = "LERIAN_MATCHER_URL"
	envMatcherClientID     = "LERIAN_MATCHER_CLIENT_ID"
	envMatcherClientSecret = "LERIAN_MATCHER_CLIENT_SECRET"
	envMatcherTokenURL     = "LERIAN_MATCHER_TOKEN_URL"
	envMatcherLegacyAPIKey = "LERIAN_MATCHER_API_KEY"

	// Tracer
	envTracerURL          = "LERIAN_TRACER_URL"
	envTracerClientID     = "LERIAN_TRACER_CLIENT_ID"
	envTracerClientSecret = "LERIAN_TRACER_CLIENT_SECRET"
	envTracerTokenURL     = "LERIAN_TRACER_TOKEN_URL"
	envTracerLegacyAPIKey = "LERIAN_TRACER_API_KEY"

	// Reporter
	envReporterURL             = "LERIAN_REPORTER_URL"
	envReporterClientID        = "LERIAN_REPORTER_CLIENT_ID"
	envReporterClientSecret    = "LERIAN_REPORTER_CLIENT_SECRET"
	envReporterTokenURL        = "LERIAN_REPORTER_TOKEN_URL"
	envReporterOrgID           = "LERIAN_REPORTER_ORG_ID"
	envReporterLegacyAuthToken = "LERIAN_REPORTER_AUTH_TOKEN"

	// Fees
	envFeesURL             = "LERIAN_FEES_URL"
	envFeesClientID        = "LERIAN_FEES_CLIENT_ID"
	envFeesClientSecret    = "LERIAN_FEES_CLIENT_SECRET"
	envFeesTokenURL        = "LERIAN_FEES_TOKEN_URL"
	envFeesOrgID           = "LERIAN_FEES_ORG_ID"
	envFeesLegacyAuthToken = "LERIAN_FEES_AUTH_TOKEN"
)

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

// envOrDefault returns the value of the environment variable named by key,
// or defaultValue if the variable is not set or empty.
func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return defaultValue
}

// envBool returns true if the environment variable named by key is set to
// "true" or "1". All other values (including empty/unset) return false.
func envBool(key string) bool {
	v := os.Getenv(key)
	return v == "true" || v == "1"
}

func anyEnvSet(keys ...string) bool {
	for _, key := range keys {
		if os.Getenv(key) != "" {
			return true
		}
	}

	return false
}

func loadMidazConfigFromEnv() *midaz.Config {
	if !anyEnvSet(
		envMidazOnboardingURL,
		envMidazTransactionURL,
		envMidazCRMURL,
		envMidazClientID,
		envMidazClientSecret,
		envMidazTokenURL,
		envMidazLegacyAuthToken,
	) {
		return nil
	}

	return &midaz.Config{
		OnboardingURL:  envOrDefault(envMidazOnboardingURL, ""),
		TransactionURL: envOrDefault(envMidazTransactionURL, ""),
		CRMURL:         envOrDefault(envMidazCRMURL, ""),
		ClientID:       envOrDefault(envMidazClientID, ""),
		ClientSecret:   envOrDefault(envMidazClientSecret, ""),
		TokenURL:       envOrDefault(envMidazTokenURL, ""),
	}
}

func loadMatcherConfigFromEnv() *matcher.Config {
	if !anyEnvSet(
		envMatcherURL,
		envMatcherClientID,
		envMatcherClientSecret,
		envMatcherTokenURL,
		envMatcherLegacyAPIKey,
	) {
		return nil
	}

	return &matcher.Config{
		BaseURL:      envOrDefault(envMatcherURL, ""),
		ClientID:     envOrDefault(envMatcherClientID, ""),
		ClientSecret: envOrDefault(envMatcherClientSecret, ""),
		TokenURL:     envOrDefault(envMatcherTokenURL, ""),
	}
}

func loadTracerConfigFromEnv() *tracer.Config {
	if !anyEnvSet(
		envTracerURL,
		envTracerClientID,
		envTracerClientSecret,
		envTracerTokenURL,
		envTracerLegacyAPIKey,
	) {
		return nil
	}

	return &tracer.Config{
		BaseURL:      envOrDefault(envTracerURL, ""),
		ClientID:     envOrDefault(envTracerClientID, ""),
		ClientSecret: envOrDefault(envTracerClientSecret, ""),
		TokenURL:     envOrDefault(envTracerTokenURL, ""),
	}
}

func loadReporterConfigFromEnv() *reporter.Config {
	if !anyEnvSet(
		envReporterURL,
		envReporterOrgID,
		envReporterClientID,
		envReporterClientSecret,
		envReporterTokenURL,
		envReporterLegacyAuthToken,
	) {
		return nil
	}

	return &reporter.Config{
		BaseURL:        envOrDefault(envReporterURL, ""),
		OrganizationID: envOrDefault(envReporterOrgID, ""),
		ClientID:       envOrDefault(envReporterClientID, ""),
		ClientSecret:   envOrDefault(envReporterClientSecret, ""),
		TokenURL:       envOrDefault(envReporterTokenURL, ""),
	}
}

func loadFeesConfigFromEnv() *fees.Config {
	if !anyEnvSet(
		envFeesURL,
		envFeesOrgID,
		envFeesClientID,
		envFeesClientSecret,
		envFeesTokenURL,
		envFeesLegacyAuthToken,
	) {
		return nil
	}

	return &fees.Config{
		BaseURL:        envOrDefault(envFeesURL, ""),
		OrganizationID: envOrDefault(envFeesOrgID, ""),
		ClientID:       envOrDefault(envFeesClientID, ""),
		ClientSecret:   envOrDefault(envFeesClientSecret, ""),
		TokenURL:       envOrDefault(envFeesTokenURL, ""),
	}
}
