package midaz

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewClient — struct construction
// ---------------------------------------------------------------------------

func TestNewClientConstruction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OnboardingURL:  "http://localhost:3000/v1",
		TransactionURL: "http://localhost:3001/v1",
		Timeout:        30 * time.Second,
	}

	// nil backends are valid for construction — services are wired separately.
	client := NewClient(nil, nil, cfg)
	require.NotNil(t, client, "NewClient must return a non-nil *Client")

	assert.Equal(t, cfg.OnboardingURL, client.config.OnboardingURL)
	assert.Equal(t, cfg.TransactionURL, client.config.TransactionURL)
	assert.Equal(t, cfg.Timeout, client.config.Timeout)
}

// ---------------------------------------------------------------------------
// NewClient — nil services do not panic
// ---------------------------------------------------------------------------

func TestNewClientWiredServicesNotNil(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OnboardingURL:  "http://localhost:3000/v1",
		TransactionURL: "http://localhost:3001/v1",
	}

	client := NewClient(nil, nil, cfg)
	require.NotNil(t, client)

	// All 13 services must be non-nil after construction.
	require.NotPanics(t, func() {
		assert.NotNil(t, client.Organizations)
		assert.NotNil(t, client.Ledgers)
		assert.NotNil(t, client.Accounts)
		assert.NotNil(t, client.AccountTypes)
		assert.NotNil(t, client.Assets)
		assert.NotNil(t, client.AssetRates)
		assert.NotNil(t, client.Balances)
		assert.NotNil(t, client.Portfolios)
		assert.NotNil(t, client.Segments)
		assert.NotNil(t, client.Transactions)
		assert.NotNil(t, client.TransactionRoutes)
		assert.NotNil(t, client.Operations)
		assert.NotNil(t, client.OperationRoutes)
	})
}

// ---------------------------------------------------------------------------
// Options — functional option pattern
// ---------------------------------------------------------------------------

func TestOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		opt       Option
		assertCfg func(*testing.T, Config)
	}{
		{
			name: "WithOnboardingURL",
			opt:  WithOnboardingURL("http://onboard:3000"),
			assertCfg: func(t *testing.T, c Config) {
				t.Helper()
				assert.Equal(t, "http://onboard:3000", c.OnboardingURL)
			},
		},
		{
			name: "WithTransactionURL",
			opt:  WithTransactionURL("http://txn:3001"),
			assertCfg: func(t *testing.T, c Config) {
				t.Helper()
				assert.Equal(t, "http://txn:3001", c.TransactionURL)
			},
		},
		{
			name: "WithClientCredentials",
			opt:  WithClientCredentials("client-id", "client-secret", "https://auth.example.com/token"),
			assertCfg: func(t *testing.T, c Config) {
				t.Helper()
				assert.Equal(t, "client-id", c.ClientID)
				assert.Equal(t, "client-secret", c.ClientSecret)
				assert.Equal(t, "https://auth.example.com/token", c.TokenURL)
			},
		},
		{
			name: "WithScopes",
			opt:  WithScopes("fees:read", "fees:write"),
			assertCfg: func(t *testing.T, c Config) {
				t.Helper()
				assert.Equal(t, []string{"fees:read", "fees:write"}, c.Scopes)
			},
		},
		{
			name: "WithTimeout",
			opt:  WithTimeout(45 * time.Second),
			assertCfg: func(t *testing.T, c Config) {
				t.Helper()
				assert.Equal(t, 45*time.Second, c.Timeout)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var cfg Config

			err := tc.opt(&cfg)
			require.NoError(t, err)
			tc.assertCfg(t, cfg)
		})
	}
}

// ---------------------------------------------------------------------------
// ErrorParser — returns the ParseError function
// ---------------------------------------------------------------------------

func TestErrorParserReturnsFunction(t *testing.T) {
	t.Parallel()

	parser := ErrorParser()
	require.NotNil(t, parser, "ErrorParser() must return a non-nil function")

	// Verify it works by parsing a simple error.
	err := parser(404, []byte(`{"code":"0040","title":"Not Found","message":"not found"}`))
	require.NotNil(t, err)
	assert.Equal(t, "midaz", err.Product)
	assert.Equal(t, 404, err.StatusCode)
}

// ---------------------------------------------------------------------------
// Service interface types — compile-time assertions
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Config credential redaction — String()
// ---------------------------------------------------------------------------

func TestConfigStringRedaction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OnboardingURL:  "http://localhost:3000/v1",
		TransactionURL: "http://localhost:3001/v1",
		ClientID:       "client-id",
		ClientSecret:   "super-secret-client-secret",
		TokenURL:       "https://auth.example.com/token",
		Scopes:         []string{"midaz:transactions"},
		Timeout:        30 * time.Second,
	}

	s := cfg.String()

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "MidazConfig")
	assert.Contains(t, s, "http://localhost:3000/v1", "OnboardingURL should be visible")
	assert.Contains(t, s, "http://localhost:3001/v1", "TransactionURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.Contains(t, s, "30s", "Timeout should be visible")
	assert.NotContains(t, s, "super-secret-client-secret",
		"String() must not contain the actual client secret")
}

// ---------------------------------------------------------------------------
// Config credential redaction — MarshalJSON()
// ---------------------------------------------------------------------------

func TestConfigMarshalJSONRedaction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OnboardingURL:  "http://localhost:3000/v1",
		TransactionURL: "http://localhost:3001/v1",
		ClientID:       "client-id",
		ClientSecret:   "super-secret-client-secret",
		TokenURL:       "https://auth.example.com/token",
		Scopes:         []string{"midaz:transactions"},
		Timeout:        30 * time.Second,
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	s := string(data)

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "http://localhost:3000/v1", "OnboardingURL should be visible")
	assert.Contains(t, s, "http://localhost:3001/v1", "TransactionURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.NotContains(t, s, "super-secret-client-secret",
		"MarshalJSON must not contain the actual client secret")
}

// ---------------------------------------------------------------------------
// Service interface types — compile-time assertions
// ---------------------------------------------------------------------------

func TestServiceInterfaceTypes(t *testing.T) {
	t.Parallel()

	// These compile-time assertions verify that each concrete entity type
	// satisfies its corresponding service interface.
	var (
		_ OrganizationsService     = (*organizationsService)(nil)
		_ LedgersService           = (*ledgersService)(nil)
		_ AccountsService          = (*accountsService)(nil)
		_ AccountTypesService      = (*accountTypesService)(nil)
		_ AssetsService            = (*assetsService)(nil)
		_ AssetRatesService        = (*assetRatesService)(nil)
		_ BalancesService          = (*balancesService)(nil)
		_ TransactionsService      = (*transactionsService)(nil)
		_ TransactionRoutesService = (*transactionRoutesService)(nil)
		_ OperationsService        = (*operationsService)(nil)
		_ OperationRoutesService   = (*operationRoutesService)(nil)
		_ PortfoliosService        = (*portfoliosService)(nil)
		_ SegmentsService          = (*segmentsService)(nil)
	)

	// The test passes if it compiles.
	t.Log("all 13 service interface types exist and compile")
}
