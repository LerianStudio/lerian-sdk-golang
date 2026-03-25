package midaz

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// Config holds the product-specific configuration for the Midaz client.
// It is populated by applying [Option] functions and validated before
// the client is constructed.
type Config struct {
	// OnboardingURL is the base URL for the Midaz onboarding API
	// (e.g. "http://localhost:3000/v1"). Required.
	OnboardingURL string

	// TransactionURL is the base URL for the Midaz transaction API
	// (e.g. "http://localhost:3001/v1"). Required.
	TransactionURL string

	// CRMURL is the base URL for the Midaz CRM API
	// (e.g. "http://localhost:4003/v1"). Required when using CRM features.
	CRMURL string

	// ClientID is the OAuth2 client identifier used for client-credentials auth.
	ClientID string

	// ClientSecret is the OAuth2 client secret used for client-credentials auth.
	ClientSecret string

	// TokenURL is the OAuth2 token endpoint URL used to acquire access tokens.
	TokenURL string

	// Timeout overrides the default HTTP client timeout for Midaz requests.
	// A zero value means the shared client timeout is used.
	Timeout time.Duration
}

// String implements fmt.Stringer to prevent credential leakage in logs.
// The ClientSecret field is replaced with "[REDACTED]".
func (c Config) String() string {
	return fmt.Sprintf(
		"MidazConfig{OnboardingURL: %q, TransactionURL: %q, CRMURL: %q, ClientID: %q, ClientSecret: [REDACTED], TokenURL: %q, Timeout: %s}",
		redactURL(c.OnboardingURL), redactURL(c.TransactionURL), redactURL(c.CRMURL), c.ClientID, redactURL(c.TokenURL), c.Timeout,
	)
}

// GoString implements fmt.GoStringer to prevent credential leakage in `%#v`
// debug formatting.
func (c Config) GoString() string {
	return c.String()
}

// MarshalJSON prevents credential leakage during JSON serialization.
// The ClientSecret field is replaced with "[REDACTED]" in the output.
func (c Config) MarshalJSON() ([]byte, error) {
	type Alias Config

	redacted := Alias(c)
	redacted.OnboardingURL = redactURL(c.OnboardingURL)
	redacted.TransactionURL = redactURL(c.TransactionURL)
	redacted.CRMURL = redactURL(c.CRMURL)
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

// Option configures a Midaz [Config]. Options are applied in order; later
// options override earlier ones.
type Option func(*Config) error

// WithOnboardingURL sets the base URL for the Midaz onboarding API.
func WithOnboardingURL(baseURL string) Option {
	return func(c *Config) error {
		c.OnboardingURL = baseURL
		return nil
	}
}

// WithTransactionURL sets the base URL for the Midaz transaction API.
func WithTransactionURL(baseURL string) Option {
	return func(c *Config) error {
		c.TransactionURL = baseURL
		return nil
	}
}

// WithClientCredentials configures OAuth2 client-credentials authentication.
// The SDK automatically acquires, caches, and refreshes access tokens using
// Lerian's client_credentials token endpoint format. Permissions are derived
// from the client configuration on the identity service; this SDK does not
// send OAuth2 scopes.
func WithClientCredentials(clientID, clientSecret, tokenURL string) Option {
	return func(c *Config) error {
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		c.TokenURL = tokenURL

		return nil
	}
}

// WithCRMURL sets the base URL for the Midaz CRM API.
func WithCRMURL(baseURL string) Option {
	return func(c *Config) error {
		c.CRMURL = baseURL
		return nil
	}
}

// WithTimeout overrides the default HTTP timeout for Midaz requests.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) error {
		c.Timeout = d
		return nil
	}
}

// ---------------------------------------------------------------------------
// Service interfaces
// ---------------------------------------------------------------------------
//
// All service interfaces are defined in their own files:
//   - organizationsServiceAPI     -> organizations.go
//   - ledgersServiceAPI           -> ledgers.go
//   - accountTypesServiceAPI      -> account_types.go
//   - accountsServiceAPI          -> accounts.go
//   - assetsServiceAPI            -> assets.go
//   - assetRatesServiceAPI        -> asset_rates.go
//   - balancesServiceAPI          -> balances.go
//   - portfoliosServiceAPI        -> portfolios.go
//   - segmentsServiceAPI          -> segments.go
//   - transactionsServiceAPI      -> transactions.go
//   - transactionRoutesServiceAPI -> transaction_routes.go
//   - operationsServiceAPI        -> operations.go
//   - operationRoutesServiceAPI   -> operation_routes.go
//   - holdersServiceAPI           -> holders.go
//   - aliasesServiceAPI           -> aliases.go

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Client is the Midaz product client. It is constructed by the root
// [lerian.Client] when the Midaz config is supplied and provides
// access to all Midaz service endpoints through domain-specific accessors.
//
// The client uses separate backends for the onboarding, transaction, and
// optional CRM APIs because Midaz splits its domain across multiple
// microservices.
type Client struct {
	// Onboarding exposes services backed by the Midaz onboarding API.
	Onboarding *OnboardingClient

	// Transactions exposes services backed by the Midaz transaction API.
	Transactions *TransactionsClient

	// CRM exposes services backed by the Midaz CRM API.
	CRM *CRMClient
}

// OnboardingClient groups all services backed by the Midaz onboarding API.
type OnboardingClient struct {
	Organizations organizationsServiceAPI
	Ledgers       ledgersServiceAPI
	Accounts      accountsServiceAPI
	AccountTypes  accountTypesServiceAPI
	Assets        assetsServiceAPI
	Portfolios    portfoliosServiceAPI
	Segments      segmentsServiceAPI
}

// TransactionsClient groups all services backed by the Midaz transaction API.
type TransactionsClient struct {
	AssetRates        assetRatesServiceAPI
	Balances          balancesServiceAPI
	Transactions      transactionsServiceAPI
	TransactionRoutes transactionRoutesServiceAPI
	Operations        operationsServiceAPI
	OperationRoutes   operationRoutesServiceAPI
}

// CRMClient groups all services backed by the Midaz CRM API.
type CRMClient struct {
	Holders holdersServiceAPI
	Aliases aliasesServiceAPI
}

// NewClient creates a Midaz product [Client] from onboarding and transaction
// backends plus a validated [Config]. Prefer constructing Midaz through the
// root [lerian.New] API unless you are wiring a custom integration layer.
func NewClient(onboardingBackend, transactionBackend core.Backend, cfg Config) *Client {
	return newClient(onboardingBackend, transactionBackend, nil, cfg)
}

// NewClientWithCRM creates a Midaz product [Client] with explicit CRM backend
// wiring in addition to onboarding and transaction backends. Prefer
// constructing Midaz through the root [lerian.New] API unless you are wiring a
// custom integration layer.
func NewClientWithCRM(onboardingBackend, transactionBackend, crmBackend core.Backend, cfg Config) *Client {
	return newClient(onboardingBackend, transactionBackend, crmBackend, cfg)
}

func newClient(onboardingBackend, transactionBackend, crmBackend core.Backend, _ Config) *Client {
	client := &Client{
		Onboarding: &OnboardingClient{
			Organizations: newOrganizationsService(onboardingBackend),
			Ledgers:       newLedgersService(onboardingBackend),
			Accounts:      newAccountsService(onboardingBackend),
			AccountTypes:  newAccountTypesService(onboardingBackend),
			Assets:        newAssetsService(onboardingBackend),
			Portfolios:    newPortfoliosService(onboardingBackend),
			Segments:      newSegmentsService(onboardingBackend),
		},
		Transactions: &TransactionsClient{
			AssetRates:        newAssetRatesService(transactionBackend),
			Balances:          newBalancesService(transactionBackend),
			Transactions:      newTransactionsService(transactionBackend),
			TransactionRoutes: newTransactionRoutesService(transactionBackend),
			Operations:        newOperationsService(transactionBackend),
			OperationRoutes:   newOperationRoutesService(transactionBackend),
		},
		CRM: &CRMClient{
			Holders: newHoldersService(crmBackend),
			Aliases: newAliasesService(crmBackend),
		},
	}

	return client
}

// ErrorParser returns the Midaz-specific error parser function suitable for
// injection into a [core.BackendConfig]. This is a convenience accessor so
// the umbrella client can wire the parser without importing internal details.
func ErrorParser() func(int, []byte) *sdkerrors.Error {
	return ParseError
}
