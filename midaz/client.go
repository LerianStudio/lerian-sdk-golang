package midaz

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

	// ClientID is the OAuth2 client identifier used for client-credentials auth.
	ClientID string

	// ClientSecret is the OAuth2 client secret used for client-credentials auth.
	ClientSecret string

	// TokenURL is the OAuth2 token endpoint URL used to acquire access tokens.
	TokenURL string

	// Scopes is the optional set of OAuth2 scopes requested during token acquisition.
	Scopes []string

	// Timeout overrides the default HTTP client timeout for Midaz requests.
	// A zero value means the shared client timeout is used.
	Timeout time.Duration
}

// String implements fmt.Stringer to prevent credential leakage in logs.
// The ClientSecret field is replaced with "[REDACTED]".
func (c Config) String() string {
	return fmt.Sprintf(
		"MidazConfig{OnboardingURL: %q, TransactionURL: %q, ClientID: %q, ClientSecret: [REDACTED], TokenURL: %q, Scopes: %v, Timeout: %s}",
		c.OnboardingURL, c.TransactionURL, c.ClientID, c.TokenURL, c.Scopes, c.Timeout,
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

// Option configures a Midaz [Config]. Options are applied in order; later
// options override earlier ones.
type Option func(*Config) error

// WithOnboardingURL sets the base URL for the Midaz onboarding API.
func WithOnboardingURL(url string) Option {
	return func(c *Config) error {
		c.OnboardingURL = url
		return nil
	}
}

// WithTransactionURL sets the base URL for the Midaz transaction API.
func WithTransactionURL(url string) Option {
	return func(c *Config) error {
		c.TransactionURL = url
		return nil
	}
}

// WithClientCredentials configures OAuth2 client-credentials authentication.
// The SDK automatically acquires, caches, and refreshes access tokens using
// the standard client_credentials grant type.
func WithClientCredentials(clientID, clientSecret, tokenURL string) Option {
	return func(c *Config) error {
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		c.TokenURL = tokenURL

		return nil
	}
}

// WithScopes sets the OAuth2 scopes requested during token acquisition.
// It is only relevant when [WithClientCredentials] is also configured.
func WithScopes(scopes ...string) Option {
	return func(c *Config) error {
		c.Scopes = append([]string(nil), scopes...)
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
//   - OrganizationsService     -> organizations.go
//   - LedgersService           -> ledgers.go
//   - AccountTypesService      -> account_types.go
//   - AccountsService          -> accounts.go
//   - AssetsService            -> assets.go
//   - AssetRatesService        -> asset_rates.go
//   - BalancesService          -> balances.go
//   - PortfoliosService        -> portfolios.go
//   - SegmentsService          -> segments.go
//   - TransactionsService      -> transactions.go
//   - TransactionRoutesService -> transaction_routes.go
//   - OperationsService        -> operations.go
//   - OperationRoutesService   -> operation_routes.go

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Client is the Midaz product client. It is constructed by the root
// [lerian.Client] when the WithMidaz option is supplied and provides
// access to all Midaz service endpoints through domain-specific accessors.
//
// The client holds two backends: one for the onboarding API and one for the
// transaction API, because Midaz splits its domain across two microservices.
type Client struct {
	onboarding  core.Backend
	transaction core.Backend
	config      Config

	// Service accessors — all 13 services are wired during construction.
	// Onboarding backend services:
	Organizations OrganizationsService
	Ledgers       LedgersService
	Accounts      AccountsService
	AccountTypes  AccountTypesService
	Assets        AssetsService
	AssetRates    AssetRatesService
	Portfolios    PortfoliosService
	Segments      SegmentsService
	// Transaction backend services:
	Balances          BalancesService
	Transactions      TransactionsService
	TransactionRoutes TransactionRoutesService
	Operations        OperationsService
	OperationRoutes   OperationRoutesService
}

// NewClient creates a Midaz product [Client] from pre-configured backends
// and a validated [Config].
//
// This function is called internally by the umbrella client during
// [lerian.New] — SDK consumers should not call it directly.
func NewClient(onboardingBackend, transactionBackend core.Backend, cfg Config) *Client {
	return &Client{
		onboarding:  onboardingBackend,
		transaction: transactionBackend,
		config:      cfg,
		// Onboarding backend services:
		Organizations: newOrganizationsService(onboardingBackend),
		Ledgers:       newLedgersService(onboardingBackend),
		Accounts:      newAccountsService(onboardingBackend),
		AccountTypes:  newAccountTypesService(onboardingBackend),
		Assets:        newAssetsService(onboardingBackend),
		AssetRates:    newAssetRatesService(onboardingBackend),
		Portfolios:    newPortfoliosService(onboardingBackend),
		Segments:      newSegmentsService(onboardingBackend),
		// Transaction backend services:
		Balances:          newBalancesService(transactionBackend),
		Transactions:      newTransactionsService(transactionBackend),
		TransactionRoutes: newTransactionRoutesService(transactionBackend),
		Operations:        newOperationsService(transactionBackend),
		OperationRoutes:   newOperationRoutesService(transactionBackend),
	}
}

// ErrorParser returns the Midaz-specific error parser function suitable for
// injection into a [core.BackendConfig]. This is a convenience accessor so
// the umbrella client can wire the parser without importing internal details.
func ErrorParser() func(int, []byte) *sdkerrors.Error {
	return ParseError
}
