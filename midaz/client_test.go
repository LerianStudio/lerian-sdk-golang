package midaz

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewClient — struct construction
// ---------------------------------------------------------------------------

func TestNewClientConstructionWithoutCRM(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OnboardingURL:  "http://localhost:3000/v1",
		TransactionURL: "http://localhost:3001/v1",
		Timeout:        30 * time.Second,
	}

	client := NewClient(nil, nil, cfg)
	require.NotNil(t, client, "NewClient must return a non-nil *Client")

	require.NotNil(t, client.CRM)
	assert.NotNil(t, client.CRM.Holders)
	assert.NotNil(t, client.CRM.Aliases)

	holder, err := client.CRM.Holders.Get(context.Background(), "org-1", "holder-1", nil)
	require.Error(t, err)
	assert.Nil(t, holder)
	assert.ErrorIs(t, err, core.ErrNilBackend)
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

	// Onboarding and transaction service accessors must be non-nil after construction.
	require.NotPanics(t, func() {
		assert.NotNil(t, client.Onboarding.Organizations)
		assert.NotNil(t, client.Onboarding.Ledgers)
		assert.NotNil(t, client.Onboarding.Accounts)
		assert.NotNil(t, client.Onboarding.AccountTypes)
		assert.NotNil(t, client.Onboarding.Assets)
		assert.NotNil(t, client.Transactions.AssetRates)
		assert.NotNil(t, client.Transactions.Balances)
		assert.NotNil(t, client.Onboarding.Portfolios)
		assert.NotNil(t, client.Onboarding.Segments)
		assert.NotNil(t, client.Transactions.Transactions)
		assert.NotNil(t, client.Transactions.TransactionRoutes)
		assert.NotNil(t, client.Transactions.Operations)
		assert.NotNil(t, client.Transactions.OperationRoutes)
		assert.NotNil(t, client.CRM.Holders)
		assert.NotNil(t, client.CRM.Aliases)
	})
}

func TestNewClientWithCRMWiresCRMServices(t *testing.T) {
	t.Parallel()

	client := NewClientWithCRM(nil, nil, &mockBackend{}, Config{CRMURL: "http://localhost:4003/v1"})
	require.NotNil(t, client)
	require.NotNil(t, client.CRM)
	assert.NotNil(t, client.CRM.Holders)
	assert.NotNil(t, client.CRM.Aliases)
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
			name: "WithCRMURL",
			opt:  WithCRMURL("http://crm:4003"),
			assertCfg: func(t *testing.T, c Config) {
				t.Helper()
				assert.Equal(t, "http://crm:4003", c.CRMURL)
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
		CRMURL:         "http://localhost:4003/v1",
		ClientID:       "client-id",
		ClientSecret:   "super-secret-client-secret",
		TokenURL:       "https://auth.example.com/token",
		Timeout:        30 * time.Second,
	}

	s := cfg.String()

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "MidazConfig")
	assert.Contains(t, s, "http://localhost:3000/v1", "OnboardingURL should be visible")
	assert.Contains(t, s, "http://localhost:3001/v1", "TransactionURL should be visible")
	assert.Contains(t, s, "http://localhost:4003/v1", "CRMURL should be visible")
	assert.Contains(t, s, "client-id", "ClientID should be visible")
	assert.Contains(t, s, "https://auth.example.com/token", "TokenURL should be visible")
	assert.Contains(t, s, "30s", "Timeout should be visible")
	assert.NotContains(t, s, "super-secret-client-secret",
		"String() must not contain the actual client secret")
}

func TestConfigGoStringRedaction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OnboardingURL:  "https://user:pass@example.com/v1",
		TransactionURL: "https://tx.example.com/v1",
		CRMURL:         "https://crm.example.com/v1",
		ClientID:       "client-id",
		ClientSecret:   "top-secret",
		TokenURL:       "https://auth.example.com/token?client_secret=super-secret",
	}

	debugValue := cfg.GoString()
	assert.NotContains(t, debugValue, "top-secret")
	assert.NotContains(t, debugValue, "pass")
	assert.NotContains(t, debugValue, "super-secret")
	assert.Contains(t, debugValue, "[REDACTED]")
}

// ---------------------------------------------------------------------------
// Config credential redaction — MarshalJSON()
// ---------------------------------------------------------------------------

func TestConfigMarshalJSONRedaction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OnboardingURL:  "http://localhost:3000/v1",
		TransactionURL: "http://localhost:3001/v1",
		CRMURL:         "http://localhost:4003/v1",
		ClientID:       "client-id",
		ClientSecret:   "super-secret-client-secret",
		TokenURL:       "https://auth.example.com/token",
		Timeout:        30 * time.Second,
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	s := string(data)

	assert.Contains(t, s, "[REDACTED]")
	assert.Contains(t, s, "http://localhost:3000/v1", "OnboardingURL should be visible")
	assert.Contains(t, s, "http://localhost:3001/v1", "TransactionURL should be visible")
	assert.Contains(t, s, "http://localhost:4003/v1", "CRMURL should be visible")
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
		_ organizationsServiceAPI     = (*organizationsService)(nil)
		_ organizationsServiceAPI     = (*organizationsService)(nil)
		_ ledgersServiceAPI           = (*ledgersService)(nil)
		_ ledgersServiceAPI           = (*ledgersService)(nil)
		_ accountsServiceAPI          = (*accountsService)(nil)
		_ accountsServiceAPI          = (*accountsService)(nil)
		_ accountTypesServiceAPI      = (*accountTypesService)(nil)
		_ assetsServiceAPI            = (*assetsService)(nil)
		_ assetsServiceAPI            = (*assetsService)(nil)
		_ assetRatesServiceAPI        = (*assetRatesService)(nil)
		_ balancesServiceAPI          = (*balancesService)(nil)
		_ balancesServiceAPI          = (*balancesService)(nil)
		_ transactionsServiceAPI      = (*transactionsService)(nil)
		_ transactionsServiceAPI      = (*transactionsService)(nil)
		_ transactionRoutesServiceAPI = (*transactionRoutesService)(nil)
		_ operationsServiceAPI        = (*operationsService)(nil)
		_ operationsServiceAPI        = (*operationsService)(nil)
		_ operationRoutesServiceAPI   = (*operationRoutesService)(nil)
		_ portfoliosServiceAPI        = (*portfoliosService)(nil)
		_ portfoliosServiceAPI        = (*portfoliosService)(nil)
		_ segmentsServiceAPI          = (*segmentsService)(nil)
		_ segmentsServiceAPI          = (*segmentsService)(nil)
		_ holdersServiceAPI           = (*holdersService)(nil)
		_ aliasesServiceAPI           = (*aliasesService)(nil)
	)

	// The test passes if it compiles.
	t.Log("all core and extension service interface types exist and compile")
}
