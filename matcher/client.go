package matcher

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// Config holds the product-specific configuration for the Matcher client.
// It is populated by applying [Option] functions and validated before
// the client is constructed.
type Config struct {
	// BaseURL is the base URL for the Matcher API
	// (e.g. "http://localhost:3002/v1"). Required.
	BaseURL string

	// ClientID is the OAuth2 client identifier used for client-credentials auth.
	ClientID string

	// ClientSecret is the OAuth2 client secret used for client-credentials auth.
	ClientSecret string

	// TokenURL is the OAuth2 token endpoint URL used to acquire access tokens.
	TokenURL string

	// Scopes is the optional set of OAuth2 scopes requested during token acquisition.
	Scopes []string

	// Timeout overrides the default HTTP client timeout for Matcher requests.
	// A zero value means the shared client timeout is used.
	Timeout time.Duration
}

// String implements fmt.Stringer to prevent credential leakage in logs.
// The ClientSecret field is replaced with "[REDACTED]".
func (c Config) String() string {
	return fmt.Sprintf(
		"MatcherConfig{BaseURL: %q, ClientID: %q, ClientSecret: [REDACTED], TokenURL: %q, Scopes: %v, Timeout: %s}",
		c.BaseURL, c.ClientID, c.TokenURL, c.Scopes, c.Timeout,
	)
}

// MarshalJSON prevents credential leakage during JSON serialization.
// The ClientSecret field is replaced with "[REDACTED]" in the output.
func (c Config) MarshalJSON() ([]byte, error) {
	type Alias Config

	return json.Marshal(&struct {
		Alias
		ClientSecret string `json:"ClientSecret"`
	}{
		Alias:        Alias(c),
		ClientSecret: "[REDACTED]",
	})
}

// Option configures a Matcher [Config]. Options are applied in order; later
// options override earlier ones.
type Option func(*Config) error

// WithBaseURL sets the base URL for the Matcher API.
func WithBaseURL(url string) Option {
	return func(c *Config) error {
		c.BaseURL = url
		return nil
	}
}

// WithClientCredentials configures OAuth2 client-credentials authentication.
func WithClientCredentials(clientID, clientSecret, tokenURL string) Option {
	return func(c *Config) error {
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		c.TokenURL = tokenURL

		return nil
	}
}

// WithScopes sets the OAuth2 scopes requested during token acquisition.
func WithScopes(scopes ...string) Option {
	return func(c *Config) error {
		c.Scopes = append([]string(nil), scopes...)
		return nil
	}
}

// WithTimeout overrides the default HTTP timeout for Matcher requests.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) error {
		c.Timeout = d
		return nil
	}
}

// ---------------------------------------------------------------------------
// Service interfaces — placeholder declarations
// ---------------------------------------------------------------------------
//
// Interfaces for Contexts, Rules, Schedules, Sources, SourceFieldMaps,
// FeeSchedules, and FieldMaps are defined in their respective service files
// (contexts.go, rules.go, schedules.go, sources.go, source_field_maps.go,
// fee_schedules.go, field_maps.go). The remaining interfaces below are
// placeholders that will be populated when the corresponding service
// implementations are built.

// ExportJobsService is defined in export_jobs.go with full method signatures.
// DisputesService is defined in disputes.go with full method signatures.
// ExceptionsService is defined in exceptions.go with full method signatures.
// GovernanceService is defined in governance.go with full method signatures.
// ImportsService is defined in imports.go with full method signatures.
// MatchingService is defined in matching.go with full method signatures.
// ReportsService is defined in reports.go with full method signatures.

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Client is the Matcher product client. It is constructed by the root
// [lerian.Client] when the WithMatcher option is supplied and provides
// access to all Matcher service endpoints through domain-specific accessors.
type Client struct {
	backend core.Backend
	config  Config

	// Service accessors — nil until the corresponding service layer is wired.
	Contexts        ContextsService
	Rules           RulesService
	Schedules       SchedulesService
	Sources         SourcesService
	SourceFieldMaps SourceFieldMapsService
	FeeSchedules    FeeSchedulesService
	FieldMaps       FieldMapsService
	ExportJobs      ExportJobsService
	Disputes        DisputesService
	Exceptions      ExceptionsService
	Governance      GovernanceService
	Imports         ImportsService
	Matching        MatchingService
	Reports         ReportsService
}

// NewClient creates a Matcher product [Client] from a pre-configured backend
// and a validated [Config].
//
// This function is called internally by the umbrella client during
// [lerian.New] — SDK consumers should not call it directly.
func NewClient(backend core.Backend, cfg Config) *Client {
	return &Client{
		backend:         backend,
		config:          cfg,
		Contexts:        newContextsService(backend),
		Rules:           newRulesService(backend),
		Schedules:       newSchedulesService(backend),
		Sources:         newSourcesService(backend),
		SourceFieldMaps: newSourceFieldMapsService(backend),
		FeeSchedules:    newFeeSchedulesService(backend),
		FieldMaps:       newFieldMapsService(backend),
		ExportJobs:      newExportJobsService(backend),
		Disputes:        newDisputesService(backend),
		Exceptions:      newExceptionsService(backend),
		Reports:         newReportsService(backend),
		Governance:      newGovernanceService(backend),
		Imports:         newImportsService(backend),
		Matching:        newMatchingService(backend),
	}
}

// ErrorParser returns the Matcher-specific error parser function suitable for
// injection into a [core.BackendConfig]. This is a convenience accessor so
// the umbrella client can wire the parser without importing internal details.
func ErrorParser() func(int, []byte) *sdkerrors.Error {
	return ParseError
}
