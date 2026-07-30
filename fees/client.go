package fees

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
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

	// ClientID is the OAuth2 client identifier used for client-credentials auth.
	ClientID string

	// ClientSecret is the OAuth2 client secret used for client-credentials auth.
	ClientSecret string

	// TokenURL is the OAuth2 token endpoint URL used to acquire access tokens.
	TokenURL string

	// OrganizationID is the organization scope for all Fees operations.
	// It is sent as the X-Organization-Id header on every request. Required.
	OrganizationID string

	// Timeout overrides the default HTTP client timeout for Fees requests.
	// A zero value means the shared client timeout is used.
	Timeout time.Duration
}

// String implements fmt.Stringer to prevent credential leakage in logs.
// The ClientSecret field is replaced with "[REDACTED]".
func (c Config) String() string {
	return fmt.Sprintf(
		"FeesConfig{BaseURL: %q, ClientID: %q, ClientSecret: [REDACTED], TokenURL: %q, OrganizationID: %q, Timeout: %s}",
		redactURL(c.BaseURL), c.ClientID, redactURL(c.TokenURL), c.OrganizationID, c.Timeout,
	)
}

// MarshalJSON prevents credential leakage during JSON serialization.
// The ClientSecret field is replaced with "[REDACTED]" in the output.
func (c Config) MarshalJSON() ([]byte, error) {
	type Alias Config

	redacted := Alias(c)
	redacted.BaseURL = redactURL(c.BaseURL)
	redacted.TokenURL = redactURL(c.TokenURL)

	return json.Marshal(&struct {
		Alias
		ClientSecret string `json:"ClientSecret"`
	}{
		Alias:        redacted,
		ClientSecret: "[REDACTED]",
	})
}

func redactURL(raw string) string {
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "[REDACTED_INVALID_URL]"
	}

	if parsed.User != nil {
		username := parsed.User.Username()
		if username != "" {
			parsed.User = url.User(username)
		} else {
			parsed.User = nil
		}
	}

	query := parsed.Query()
	for key := range query {
		if isSensitiveURLQueryKey(key) {
			query.Set(key, "[REDACTED]")
		}
	}

	parsed.RawQuery = query.Encode()
	if parsed.Fragment != "" {
		parsed.Fragment = "[REDACTED]"
	}

	return parsed.String()
}

func isSensitiveURLQueryKey(key string) bool {
	normalized := strings.ToLower(key)
	for _, token := range []string{"token", "secret", "password", "signature", "sig", "key", "auth"} {
		if strings.Contains(normalized, token) {
			return true
		}
	}

	return false
}

// Option configures a Fees [Config]. Options are applied in order; later
// options override earlier ones.
type Option func(*Config) error

// WithBaseURL sets the base URL for the Fees API.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) error {
		c.BaseURL = baseURL
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
//   - packagesServiceAPI  → packages.go
//   - estimatesServiceAPI → estimates.go
//   - feesServiceAPI      → fees_service.go

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Client is the Fees product client. It is typically constructed by the root
// [lerian.Client] from [lerian.Config] and provides access to all Fees service
// endpoints through domain-specific accessors.
type Client struct {
	// Packages provides access to fee package management endpoints.
	Packages packagesServiceAPI

	// Estimates provides access to fee estimation (preview) endpoints.
	Estimates estimatesServiceAPI

	// Fees provides access to fee calculation and management endpoints.
	Fees feesServiceAPI
}

// NewClient creates a Fees product [Client] from a pre-configured backend
// and a validated [Config].
//
// Prefer constructing Fees through the root [lerian.New] API unless you are
// wiring a custom integration layer.
func NewClient(backend core.Backend, _ Config) *Client {
	return &Client{
		Packages:  newPackagesService(backend),
		Estimates: newEstimatesService(backend),
		Fees:      newFeesCalcService(backend),
	}
}
