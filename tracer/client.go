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

// defaultTimeout is the default HTTP timeout for Tracer requests.
// Tracer operations tend to be fast lookups, so a shorter timeout is used.
const defaultTimeout = 10 * time.Second

// Config holds the product-specific configuration for the Tracer client.
// It is populated by applying [Option] functions and validated before
// the client is constructed.
type Config struct {
	// BaseURL is the base URL for the Tracer API
	// (e.g. "http://localhost:3003/v1"). Required.
	BaseURL string

	// APIKey is the API key used to authenticate Tracer requests.
	// It is sent in the X-API-Key header.
	APIKey string

	// Timeout overrides the default HTTP client timeout for Tracer requests.
	// Defaults to 10s if zero.
	Timeout time.Duration
}

// String implements fmt.Stringer to prevent credential leakage in logs.
// The APIKey field is replaced with "[REDACTED]".
func (c Config) String() string {
	return fmt.Sprintf(
		"TracerConfig{BaseURL: %q, APIKey: [REDACTED], Timeout: %s}",
		c.BaseURL, c.Timeout,
	)
}

// MarshalJSON prevents credential leakage during JSON serialization.
// The APIKey field is replaced with "[REDACTED]" in the output.
func (c Config) MarshalJSON() ([]byte, error) {
	type Alias Config

	return json.Marshal(&struct {
		Alias
		APIKey string `json:"APIKey"`
	}{
		Alias:  Alias(c),
		APIKey: "[REDACTED]",
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

// WithAPIKey sets the API key for Tracer authentication.
func WithAPIKey(key string) Option {
	return func(c *Config) error {
		c.APIKey = key
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

// Client is the Tracer product client. It is constructed by the root
// [lerian.Client] when the WithTracer option is supplied and provides
// access to all Tracer service endpoints through domain-specific accessors.
type Client struct {
	backend core.Backend
	config  Config

	// Rules provides access to compliance rule management endpoints.
	Rules RulesService

	// Limits provides access to rate/amount limit management endpoints.
	Limits LimitsService

	// Validations provides access to transaction validation endpoints.
	Validations ValidationsService

	// AuditEvents provides access to audit trail query and verification endpoints.
	AuditEvents AuditEventsService
}

// NewClient creates a Tracer product [Client] from a pre-configured backend
// and a validated [Config].
//
// This function is called internally by the umbrella client during
// [lerian.New] -- SDK consumers should not call it directly.
func NewClient(backend core.Backend, cfg Config) *Client {
	return &Client{
		config:      cfg,
		backend:     backend,
		Rules:       newRulesService(backend),
		Limits:      newLimitsService(backend),
		Validations: newValidationsService(backend),
		AuditEvents: newAuditEventsService(backend),
	}
}
