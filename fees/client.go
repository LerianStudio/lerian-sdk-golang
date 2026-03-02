package fees

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// Config holds the product-specific configuration for the Fees client.
// It is populated by applying [Option] functions and validated before
// the client is constructed.
type Config struct {
	// BaseURL is the base URL for the Fees API
	// (e.g. "http://localhost:3005/v1"). Required.
	BaseURL string

	// AuthToken is the bearer token used to authenticate Fees requests.
	AuthToken string

	// OrganizationID is the organization scope for all Fees operations.
	// It is sent as the X-Organization-Id header on every request. Required.
	OrganizationID string

	// Timeout overrides the default HTTP client timeout for Fees requests.
	// A zero value means the shared client timeout is used.
	Timeout time.Duration
}

// String implements fmt.Stringer to prevent credential leakage in logs.
// The AuthToken field is replaced with "[REDACTED]".
func (c Config) String() string {
	return fmt.Sprintf(
		"FeesConfig{BaseURL: %q, AuthToken: [REDACTED], OrganizationID: %q, Timeout: %s}",
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

// Option configures a Fees [Config]. Options are applied in order; later
// options override earlier ones.
type Option func(*Config) error

// WithBaseURL sets the base URL for the Fees API.
func WithBaseURL(url string) Option {
	return func(c *Config) error {
		c.BaseURL = url
		return nil
	}
}

// WithAuthToken sets the bearer token for Fees API authentication.
func WithAuthToken(token string) Option {
	return func(c *Config) error {
		c.AuthToken = token
		return nil
	}
}

// WithOrganizationID sets the organization scope for Fees operations.
func WithOrganizationID(orgID string) Option {
	return func(c *Config) error {
		c.OrganizationID = orgID
		return nil
	}
}

// WithTimeout overrides the default HTTP timeout for Fees requests.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) error {
		c.Timeout = d
		return nil
	}
}

// Service interfaces are declared in their respective service files:
//   - PackagesService  → packages.go
//   - EstimatesService → estimates.go
//   - FeesService      → fees_service.go

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Client is the Fees product client. It is constructed by the root
// [lerian.Client] when the WithFees option is supplied and provides
// access to all Fees service endpoints through domain-specific accessors.
type Client struct {
	backend core.Backend
	config  Config

	// Packages provides access to fee package management endpoints.
	Packages PackagesService

	// Estimates provides access to fee estimation (preview) endpoints.
	Estimates EstimatesService

	// Fees provides access to fee calculation and management endpoints.
	Fees FeesService
}

// NewClient creates a Fees product [Client] from a pre-configured backend
// and a validated [Config].
//
// This function is called internally by the umbrella client during
// [lerian.New] -- SDK consumers should not call it directly.
func NewClient(backend core.Backend, cfg Config) *Client {
	return &Client{
		backend:   backend,
		config:    cfg,
		Packages:  newPackagesService(backend),
		Estimates: newEstimatesService(backend),
		Fees:      newFeesCalcService(backend),
	}
}
