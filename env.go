package lerian

import "os"

// ---------------------------------------------------------------------------
// Environment variable constants
// ---------------------------------------------------------------------------

// LERIAN_* environment variables provide fallback configuration when explicit
// options are not supplied. The precedence order is:
//
//	explicit option > environment variable > empty (fails validation if required)
//
// Only the LERIAN_* prefix is recognized; legacy MIDAZ_* variables from v2
// are deliberately ignored to enforce a clean break.
const (
	// Debug
	envDebug = "LERIAN_DEBUG"

	// Midaz
	envMidazOnboardingURL   = "LERIAN_MIDAZ_ONBOARDING_URL"
	envMidazTransactionURL  = "LERIAN_MIDAZ_TRANSACTION_URL"
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
