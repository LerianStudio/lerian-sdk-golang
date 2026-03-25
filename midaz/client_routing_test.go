package midaz

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientRoutesAssetRatesToTransactionBackend(t *testing.T) {
	t.Parallel()

	onboardingCalled := false
	transactionCalled := false

	onboarding := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		onboardingCalled = true

		return nil
	}}
	transaction := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
		transactionCalled = true

		assert.Equal(t, "GET", method)
		assert.Equal(t, "/organizations/org-1/ledgers/led-1/asset-rates/rate-1", path)
		assert.Nil(t, body)

		return unmarshalInto(AssetRate{ID: "rate-1"}, result)
	}}

	client := NewClient(onboarding, transaction, Config{})
	rate, err := client.Transactions.AssetRates.Get(context.Background(), "org-1", "led-1", "rate-1")

	require.NoError(t, err)
	require.NotNil(t, rate)
	assert.False(t, onboardingCalled)
	assert.True(t, transactionCalled)
}

func TestClientRoutesOrganizationsToOnboardingBackend(t *testing.T) {
	t.Parallel()

	onboardingCalled := false
	transactionCalled := false

	onboarding := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
		onboardingCalled = true

		assert.Equal(t, "GET", method)
		assert.Equal(t, "/organizations/org-1", path)
		assert.Nil(t, body)

		return unmarshalInto(Organization{ID: "org-1"}, result)
	}}
	transaction := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		transactionCalled = true
		return nil
	}}

	client := NewClient(onboarding, transaction, Config{})
	org, err := client.Onboarding.Organizations.Get(context.Background(), "org-1")

	require.NoError(t, err)
	require.NotNil(t, org)
	assert.True(t, onboardingCalled)
	assert.False(t, transactionCalled)
}

func TestClientRoutesCRMServicesToCRMBackend(t *testing.T) {
	t.Parallel()

	onboardingCalled := false
	transactionCalled := false
	crmCalled := false

	onboarding := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		onboardingCalled = true

		return nil
	}}
	transaction := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		transactionCalled = true

		return nil
	}}
	crm := &mockBackend{callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
		crmCalled = true

		assert.Equal(t, "GET", method)
		assert.Equal(t, "/holders/holder-1", path)
		assert.Equal(t, "org-1", headers["X-Organization-Id"])

		return unmarshalInto(Holder{ID: "holder-1"}, result)
	}}

	client := NewClientWithCRM(onboarding, transaction, crm, Config{})
	holder, err := client.CRM.Holders.Get(context.Background(), "org-1", "holder-1", nil)

	require.NoError(t, err)
	require.NotNil(t, holder)
	assert.False(t, onboardingCalled)
	assert.False(t, transactionCalled)
	assert.True(t, crmCalled)
}

func TestClientRoutesAliasesToCRMBackend(t *testing.T) {
	t.Parallel()

	onboardingCalled := false
	transactionCalled := false
	crmCalled := false

	onboarding := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		onboardingCalled = true
		return nil
	}}
	transaction := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		transactionCalled = true
		return nil
	}}
	crm := &mockBackend{callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
		crmCalled = true

		assert.Equal(t, "GET", method)
		assert.Equal(t, "/holders/holder-1/aliases/alias-1", path)
		assert.Equal(t, "org-1", headers["X-Organization-Id"])

		return unmarshalInto(Alias{ID: "alias-1", HolderID: "holder-1"}, result)
	}}

	client := NewClientWithCRM(onboarding, transaction, crm, Config{})
	alias, err := client.CRM.Aliases.Get(context.Background(), "org-1", "holder-1", "alias-1", nil)
	require.NoError(t, err)
	require.NotNil(t, alias)
	assert.False(t, onboardingCalled)
	assert.False(t, transactionCalled)
	assert.True(t, crmCalled)
}

func TestClientRoutesBalancesToTransactionBackend(t *testing.T) {
	t.Parallel()

	onboardingCalled := false
	transactionCalled := false

	onboarding := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		onboardingCalled = true
		return nil
	}}
	transaction := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
		transactionCalled = true

		assert.Equal(t, "GET", method)
		assert.Equal(t, "/organizations/org-1/ledgers/led-1/accounts/account-1/balances", path)
		assert.Nil(t, body)

		return unmarshalInto(balancesLookupResponse{Items: []Balance{{ID: "bal-1"}}}, result)
	}}

	client := NewClient(onboarding, transaction, Config{})
	items, err := client.Transactions.Balances.ListByAccountID(context.Background(), "org-1", "led-1", "account-1")
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.False(t, onboardingCalled)
	assert.True(t, transactionCalled)
}

func TestClientRoutesTransactionsToTransactionBackend(t *testing.T) {
	t.Parallel()

	onboardingCalled := false
	transactionCalled := false

	onboarding := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		onboardingCalled = true
		return nil
	}}
	transaction := &mockBackend{callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
		transactionCalled = true

		assert.Equal(t, "POST", method)
		assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/dsl", path)
		assert.Contains(t, headers["Content-Type"], "multipart/form-data")
		assert.IsType(t, []byte{}, body)

		return unmarshalInto(Transaction{ID: "txn-dsl"}, result)
	}}

	client := NewClient(onboarding, transaction, Config{})
	txn, err := client.Transactions.Transactions.CreateDSL(context.Background(), "org-1", "led-1", []byte("SEND 1 BRL"))
	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.False(t, onboardingCalled)
	assert.True(t, transactionCalled)
}

func TestClientRoutesOperationsToTransactionBackend(t *testing.T) {
	t.Parallel()

	onboardingCalled := false
	transactionCalled := false

	onboarding := &mockBackend{callFn: func(_ context.Context, _, _ string, _, _ any) error {
		onboardingCalled = true
		return nil
	}}
	transaction := &mockBackend{callFn: func(_ context.Context, method, path string, body, result any) error {
		transactionCalled = true

		assert.Equal(t, "PATCH", method)
		assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/tx-1/operations/op-1", path)
		assert.NotNil(t, body)

		return unmarshalInto(Operation{ID: "op-1", TransactionID: "tx-1"}, result)
	}}

	client := NewClient(onboarding, transaction, Config{})
	op, err := client.Transactions.Operations.Update(context.Background(), "org-1", "led-1", "tx-1", "op-1", &UpdateOperationInput{})
	require.NoError(t, err)
	require.NotNil(t, op)
	assert.False(t, onboardingCalled)
	assert.True(t, transactionCalled)
}

func TestConfigURLRedaction(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OnboardingURL:  "https://user:pass@example.com/v1?token=secret-token&mode=test",
		TransactionURL: "https://tx.example.com/v1?api_key=shh",
		CRMURL:         "https://crm.example.com/v1?signature=abc123",
		TokenURL:       "https://auth.example.com/token?client_secret=top-secret",
		ClientSecret:   "client-secret",
	}

	stringValue := cfg.String()
	assert.NotContains(t, stringValue, "pass")
	assert.NotContains(t, stringValue, "secret-token")
	assert.NotContains(t, stringValue, "abc123")
	assert.NotContains(t, stringValue, "top-secret")
	assert.Contains(t, stringValue, "mode=test")

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	serialized := string(data)
	assert.NotContains(t, serialized, "pass")
	assert.NotContains(t, serialized, "secret-token")
	assert.NotContains(t, serialized, "abc123")
	assert.NotContains(t, serialized, "top-secret")
	assert.Contains(t, serialized, "[REDACTED]")
}

func TestConfigURLRedactionHandlesInvalidURLsAndFragments(t *testing.T) {
	t.Parallel()

	cfg := Config{
		OnboardingURL: "https://example.com/callback#access_token=secret-token",
		TokenURL:      "https://example.com/%zz?token=secret-token",
		ClientSecret:  "client-secret",
	}

	stringValue := cfg.String()
	assert.NotContains(t, stringValue, "secret-token")
	assert.Contains(t, stringValue, "[REDACTED]")
	assert.Contains(t, stringValue, "[REDACTED_INVALID_URL]")

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	serialized := string(data)
	assert.NotContains(t, serialized, "secret-token")
	assert.Contains(t, serialized, "[REDACTED]")
	assert.Contains(t, serialized, "[REDACTED_INVALID_URL]")
}
