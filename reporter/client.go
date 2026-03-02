package reporter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// Config holds the product-specific configuration for the Reporter client.
// It is populated by applying [Option] functions and validated before
// the client is constructed.
type Config struct {
	// BaseURL is the base URL for the Reporter API
	// (e.g. "http://localhost:3004/v1"). Required.
	BaseURL string

	// AuthToken is the bearer token used to authenticate Reporter requests.
	AuthToken string

	// OrganizationID is the organization scope for all Reporter operations.
	// It is sent as the X-Organization-Id header on every request. Required.
	OrganizationID string

	// Timeout overrides the default HTTP client timeout for Reporter requests.
	// A zero value means the shared client timeout is used.
	Timeout time.Duration
}

// String implements fmt.Stringer to prevent credential leakage in logs.
// The AuthToken field is replaced with "[REDACTED]".
func (c Config) String() string {
	return fmt.Sprintf(
		"ReporterConfig{BaseURL: %q, AuthToken: [REDACTED], OrganizationID: %q, Timeout: %s}",
		c.BaseURL, c.OrganizationID, c.Timeout,
	)
}

// MarshalJSON prevents credential leakage during JSON serialization.
// The AuthToken field is replaced with "[REDACTED]" in the output.
func (c Config) MarshalJSON() ([]byte, error) {
	type Alias Config

	return json.Marshal(&struct {
		Alias
		AuthToken string `json:"AuthToken"`
	}{
		Alias:     Alias(c),
		AuthToken: "[REDACTED]",
	})
}

// Option configures a Reporter [Config]. Options are applied in order; later
// options override earlier ones.
type Option func(*Config) error

// WithBaseURL sets the base URL for the Reporter API.
func WithBaseURL(url string) Option {
	return func(c *Config) error {
		c.BaseURL = url
		return nil
	}
}

// WithAuthToken sets the bearer token for Reporter API authentication.
func WithAuthToken(token string) Option {
	return func(c *Config) error {
		c.AuthToken = token
		return nil
	}
}

// WithOrganizationID sets the organization scope for Reporter operations.
func WithOrganizationID(orgID string) Option {
	return func(c *Config) error {
		c.OrganizationID = orgID
		return nil
	}
}

// WithTimeout overrides the default HTTP timeout for Reporter requests.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) error {
		c.Timeout = d
		return nil
	}
}

// ---------------------------------------------------------------------------
// Service interfaces are defined in their respective files:
//   - DataSourcesService  → datasources.go
//   - ReportsService      → reports.go
//   - TemplatesService    → templates.go
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Client is the Reporter product client. It is constructed by the root
// [lerian.Client] when the WithReporter option is supplied and provides
// access to all Reporter service endpoints through domain-specific accessors.
type Client struct {
	backend core.Backend
	config  Config

	// Service accessors for Reporter domain endpoints.
	DataSources DataSourcesService
	Reports     ReportsService
	Templates   TemplatesService
}

// NewClient creates a Reporter product [Client] from a pre-configured backend
// and a validated [Config].
//
// This function is called internally by the umbrella client during
// [lerian.New] — SDK consumers should not call it directly.
func NewClient(backend core.Backend, cfg Config) *Client {
	return &Client{
		backend:     backend,
		config:      cfg,
		DataSources: newDataSourcesService(backend),
		Reports:     newReportsService(backend),
		Templates:   newTemplatesService(backend),
	}
}
