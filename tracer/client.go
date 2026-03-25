package tracer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// Config holds the product-specific configuration for the Tracer client.
// It is populated by applying [Option] functions and validated before
// the client is constructed.
type Config struct {
	// BaseURL is the base URL for the Tracer API
	// (e.g. "http://localhost:3003/v1"). Required.
	BaseURL string

	// ClientID is the OAuth2 client identifier used for client-credentials auth.
	ClientID string

	// ClientSecret is the OAuth2 client secret used for client-credentials auth.
	ClientSecret string

	// TokenURL is the OAuth2 token endpoint URL used to acquire access tokens.
	TokenURL string

	// Timeout overrides the default HTTP client timeout for Tracer requests.
	// Defaults to 10s if zero.
	Timeout time.Duration
}

// String implements fmt.Stringer to prevent credential leakage in logs.
// The ClientSecret field is replaced with "[REDACTED]".
func (c Config) String() string {
	return fmt.Sprintf(
		"TracerConfig{BaseURL: %q, ClientID: %q, ClientSecret: [REDACTED], TokenURL: %q, Timeout: %s}",
		c.BaseURL, c.ClientID, c.TokenURL, c.Timeout,
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

// Option configures a Tracer [Config]. Options are applied in order; later
// options override earlier ones.
type Option func(*Config) error

// WithBaseURL sets the base URL for the Tracer API.
func WithBaseURL(url string) Option {
	return func(c *Config) error {
		c.BaseURL = url
		return nil
	}
}

// WithClientCredentials configures OAuth2 client-credentials authentication.
// Permissions are derived from the client configuration on the identity
// service; this SDK does not send OAuth2 scopes.
func WithClientCredentials(clientID, clientSecret, tokenURL string) Option {
	return func(c *Config) error {
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		c.TokenURL = tokenURL

		return nil
	}
}

// WithTimeout overrides the default HTTP timeout for Tracer requests.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) error {
		c.Timeout = d
		return nil
	}
}

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Client is the Tracer product client. It is typically constructed by the root
// [lerian.Client] from [lerian.Config] and provides access to all Tracer
// service endpoints through domain-specific accessors.
type Client struct {
	// Rules provides access to compliance rule management endpoints.
	Rules rulesServiceAPI

	// Limits provides access to rate/amount limit management endpoints.
	Limits limitsServiceAPI

	// Validations provides access to transaction validation endpoints.
	Validations validationsServiceAPI

	// AuditEvents provides access to audit trail query and verification endpoints.
	AuditEvents auditEventsServiceAPI
}

// NewClient creates a Tracer product [Client] from a pre-configured backend
// and a validated [Config].
//
// Prefer constructing Tracer through the root [lerian.New] API unless you are
// wiring a custom integration layer.
func NewClient(backend core.Backend, _ Config) *Client {
	return &Client{
		Rules:       newRulesService(backend),
		Limits:      newLimitsService(backend),
		Validations: newValidationsService(backend),
		AuditEvents: newAuditEventsService(backend),
	}
}
