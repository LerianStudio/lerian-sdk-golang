package leriantest_test

import (
	"context"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/midaz"
	"github.com/LerianStudio/lerian-sdk-golang/testing/leriantest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================================================================
// 1. AccountTypes -- full CRUD
// ===========================================================================

func TestFakeMidazAccountTypesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	orgID := "org-1"
	ledgerID := "ledger-1"
	desc := "Savings accounts"

	// Create
	created, err := client.Midaz.Onboarding.AccountTypes.Create(ctx, orgID, ledgerID, &midaz.CreateAccountTypeInput{
		Name:        "savings",
		Description: &desc,
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "savings", created.Name)
	assert.Equal(t, &desc, created.Description)
	assert.Equal(t, orgID, created.OrganizationID)
	assert.Equal(t, ledgerID, created.LedgerID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Midaz.Onboarding.AccountTypes.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "savings", got.Name)

	// Update
	newName := "deposit"
	updated, err := client.Midaz.Onboarding.AccountTypes.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateAccountTypeInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "deposit", updated.Name)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Onboarding.AccountTypes.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Onboarding.AccountTypes.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Onboarding.AccountTypes.Get(ctx, orgID, ledgerID, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazAccountTypesNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-account-type-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Midaz.Onboarding.AccountTypes.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.Onboarding.AccountTypes.Update(ctx, "o", "l", ghost, &midaz.UpdateAccountTypeInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Onboarding.AccountTypes.Delete(ctx, "o", "l", ghost) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

func TestFakeMidazAccountTypesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("midaz.AccountTypes.Create", injectedErr),
	)

	_, err := client.Midaz.Onboarding.AccountTypes.Create(ctx, "o", "l", &midaz.CreateAccountTypeInput{
		Name: "Should Fail",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

// ===========================================================================
// 2. Assets -- full CRUD
// ===========================================================================

func TestFakeMidazAssetsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	orgID := "org-1"
	ledgerID := "ledger-1"

	// Create
	created, err := client.Midaz.Onboarding.Assets.Create(ctx, orgID, ledgerID, &midaz.CreateAssetInput{
		Name: "Brazilian Real",
		Code: "BRL",
		Type: "currency",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Brazilian Real", created.Name)
	assert.Equal(t, "BRL", created.Code)
	assert.Equal(t, "currency", created.Type)
	assert.Equal(t, "active", created.Status.Code)
	assert.Equal(t, orgID, created.OrganizationID)
	assert.Equal(t, ledgerID, created.LedgerID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Midaz.Onboarding.Assets.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "BRL", got.Code)

	// Update
	newName := "Real Brasileiro"
	updated, err := client.Midaz.Onboarding.Assets.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateAssetInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Real Brasileiro", updated.Name)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Onboarding.Assets.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Onboarding.Assets.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Onboarding.Assets.Get(ctx, orgID, ledgerID, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazAssetsNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-asset-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Midaz.Onboarding.Assets.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.Onboarding.Assets.Update(ctx, "o", "l", ghost, &midaz.UpdateAssetInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Onboarding.Assets.Delete(ctx, "o", "l", ghost) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

func TestFakeMidazAssetsListMultiple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	for i := 0; i < 3; i++ {
		_, err := client.Midaz.Onboarding.Assets.Create(ctx, "o", "l", &midaz.CreateAssetInput{
			Name: "Asset",
			Code: "TST",
			Type: "currency",
		})
		require.NoError(t, err)
	}

	iter := client.Midaz.Onboarding.Assets.List(ctx, "o", "l", nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 3)
}

// ===========================================================================
// 3. AssetRates -- full CRUD + GetByExternalID + GetFromAssetCode
// ===========================================================================

func TestFakeMidazAssetRatesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	orgID := "org-1"
	ledgerID := "ledger-1"
	source := "central-bank"
	externalID := "ext-rate-001"

	// Create
	created, err := client.Midaz.Transactions.AssetRates.Create(ctx, orgID, ledgerID, &midaz.CreateAssetRateInput{
		BaseAssetCode:    "BRL",
		CounterAssetCode: "USD",
		Amount:           550,
		Scale:            2,
		Source:           &source,
		ExternalID:       &externalID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "BRL", created.BaseAssetCode)
	assert.Equal(t, "USD", created.CounterAssetCode)
	assert.Equal(t, int64(550), created.Amount)
	assert.Equal(t, 2, created.Scale)
	assert.Equal(t, &source, created.Source)
	assert.Equal(t, &externalID, created.ExternalID)
	assert.Equal(t, orgID, created.OrganizationID)
	assert.Equal(t, ledgerID, created.LedgerID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Midaz.Transactions.AssetRates.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "BRL", got.BaseAssetCode)

	// Update
	newAmount := int64(560)
	newScale := 2
	updated, err := client.Midaz.Transactions.AssetRates.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateAssetRateInput{
		Amount: &newAmount,
		Scale:  &newScale,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, int64(560), updated.Amount)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Transactions.AssetRates.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Transactions.AssetRates.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Transactions.AssetRates.Get(ctx, orgID, ledgerID, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazAssetRatesNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-rate-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Midaz.Transactions.AssetRates.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.Transactions.AssetRates.Update(ctx, "o", "l", ghost, &midaz.UpdateAssetRateInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Transactions.AssetRates.Delete(ctx, "o", "l", ghost) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

func TestFakeMidazAssetRatesGetByExternalID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	externalID := "ext-rate-lookup"

	created, err := client.Midaz.Transactions.AssetRates.Create(ctx, "o", "l", &midaz.CreateAssetRateInput{
		BaseAssetCode:    "BRL",
		CounterAssetCode: "EUR",
		Amount:           600,
		Scale:            2,
		ExternalID:       &externalID,
	})
	require.NoError(t, err)

	// Found
	got, err := client.Midaz.Transactions.AssetRates.GetByExternalID(ctx, "o", "l", "ext-rate-lookup")
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Not found
	_, err = client.Midaz.Transactions.AssetRates.GetByExternalID(ctx, "o", "l", "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazAssetRatesGetFromAssetCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	created, err := client.Midaz.Transactions.AssetRates.Create(ctx, "o", "l", &midaz.CreateAssetRateInput{
		BaseAssetCode:    "GBP",
		CounterAssetCode: "USD",
		Amount:           130,
		Scale:            2,
	})
	require.NoError(t, err)

	// Found by base asset code
	got, err := client.Midaz.Transactions.AssetRates.GetFromAssetCode(ctx, "o", "l", "GBP")
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Not found
	_, err = client.Midaz.Transactions.AssetRates.GetFromAssetCode(ctx, "o", "l", "JPY")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ===========================================================================
// 4. Portfolios -- full CRUD
// ===========================================================================

func TestFakeMidazPortfoliosCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	orgID := "org-1"
	ledgerID := "ledger-1"
	entityID := "entity-abc"

	// Create
	created, err := client.Midaz.Onboarding.Portfolios.Create(ctx, orgID, ledgerID, &midaz.CreatePortfolioInput{
		Name:     "Main Portfolio",
		EntityID: &entityID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Main Portfolio", created.Name)
	assert.Equal(t, &entityID, created.EntityID)
	assert.Equal(t, "active", created.Status.Code)
	assert.Equal(t, orgID, created.OrganizationID)
	assert.Equal(t, ledgerID, created.LedgerID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Midaz.Onboarding.Portfolios.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Main Portfolio", got.Name)

	// Update
	newName := "Updated Portfolio"
	updated, err := client.Midaz.Onboarding.Portfolios.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdatePortfolioInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Updated Portfolio", updated.Name)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Onboarding.Portfolios.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Onboarding.Portfolios.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Onboarding.Portfolios.Get(ctx, orgID, ledgerID, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazPortfoliosNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-portfolio-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Midaz.Onboarding.Portfolios.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.Onboarding.Portfolios.Update(ctx, "o", "l", ghost, &midaz.UpdatePortfolioInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Onboarding.Portfolios.Delete(ctx, "o", "l", ghost) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

func TestFakeMidazPortfoliosErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("midaz.Portfolios.Create", injectedErr),
	)

	_, err := client.Midaz.Onboarding.Portfolios.Create(ctx, "o", "l", &midaz.CreatePortfolioInput{
		Name: "Should Fail",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

// ===========================================================================
// 5. Segments -- full CRUD
// ===========================================================================

func TestFakeMidazSegmentsCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	orgID := "org-1"
	ledgerID := "ledger-1"

	// Create
	created, err := client.Midaz.Onboarding.Segments.Create(ctx, orgID, ledgerID, &midaz.CreateSegmentInput{
		Name: "Retail Banking",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Retail Banking", created.Name)
	assert.Equal(t, "active", created.Status.Code)
	assert.Equal(t, orgID, created.OrganizationID)
	assert.Equal(t, ledgerID, created.LedgerID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Midaz.Onboarding.Segments.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Retail Banking", got.Name)

	// Update
	newName := "Corporate Banking"
	updated, err := client.Midaz.Onboarding.Segments.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateSegmentInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Corporate Banking", updated.Name)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Onboarding.Segments.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Onboarding.Segments.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Onboarding.Segments.Get(ctx, orgID, ledgerID, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazSegmentsNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-segment-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Midaz.Onboarding.Segments.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.Onboarding.Segments.Update(ctx, "o", "l", ghost, &midaz.UpdateSegmentInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Onboarding.Segments.Delete(ctx, "o", "l", ghost) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

func TestFakeMidazSegmentsListMultiple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	for i := 0; i < 3; i++ {
		_, err := client.Midaz.Onboarding.Segments.Create(ctx, "o", "l", &midaz.CreateSegmentInput{
			Name: "Segment",
		})
		require.NoError(t, err)
	}

	iter := client.Midaz.Onboarding.Segments.List(ctx, "o", "l", nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 3)
}

// ===========================================================================
// 6. Balances -- full CRUD + ListByAlias + ListByExternalCode + ListByAccountID
// ===========================================================================

func TestFakeMidazBalancesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	orgID := "org-1"
	ledgerID := "ledger-1"
	alias := "@acct-123"

	account, err := client.Midaz.Onboarding.Accounts.Create(ctx, orgID, ledgerID, &midaz.CreateAccountInput{
		Name:      "Account 123",
		Type:      "asset",
		AssetCode: "BRL",
		Alias:     &alias,
	})
	require.NoError(t, err)

	// Create
	created, err := client.Midaz.Transactions.Balances.CreateForAccount(ctx, orgID, ledgerID, account.ID, &midaz.CreateBalanceInput{Key: "asset-freeze"})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, account.ID, created.AccountID)
	assert.Equal(t, "BRL", created.AssetCode)
	assert.Equal(t, "active", created.Status.Code)
	assert.True(t, created.AllowSending)
	assert.True(t, created.AllowReceiving)
	assert.Equal(t, orgID, created.OrganizationID)
	assert.Equal(t, ledgerID, created.LedgerID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Midaz.Transactions.Balances.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, account.ID, got.AccountID)

	// Update
	allowSending := false
	updated, err := client.Midaz.Transactions.Balances.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateBalanceInput{
		AllowSending: &allowSending,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.False(t, updated.AllowSending)
	assert.True(t, updated.AllowReceiving) // unchanged
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Transactions.Balances.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Transactions.Balances.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Transactions.Balances.Get(ctx, orgID, ledgerID, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazBalancesNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-balance-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Midaz.Transactions.Balances.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.Transactions.Balances.Update(ctx, "o", "l", ghost, &midaz.UpdateBalanceInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Transactions.Balances.Delete(ctx, "o", "l", ghost) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

func TestFakeMidazBalancesListByAccountID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	account, err := client.Midaz.Onboarding.Accounts.Create(ctx, "o", "l", &midaz.CreateAccountInput{
		Name:      "Lookup",
		Type:      "asset",
		AssetCode: "USD",
	})
	require.NoError(t, err)

	created, err := client.Midaz.Transactions.Balances.CreateForAccount(ctx, "o", "l", account.ID, &midaz.CreateBalanceInput{Key: "asset-freeze"})
	require.NoError(t, err)

	// Found
	items, err := client.Midaz.Transactions.Balances.ListByAccountID(ctx, "o", "l", account.ID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Not found
	items, err = client.Midaz.Transactions.Balances.ListByAccountID(ctx, "o", "l", "nonexistent-acct")
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFakeMidazBalancesListByAlias(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	alias := "@balance-alias"
	account, err := client.Midaz.Onboarding.Accounts.Create(ctx, "o", "l", &midaz.CreateAccountInput{
		Name:      "Alias Lookup",
		Type:      "asset",
		AssetCode: "BRL",
		Alias:     &alias,
	})
	require.NoError(t, err)

	created, err := client.Midaz.Transactions.Balances.CreateForAccount(ctx, "o", "l", account.ID, &midaz.CreateBalanceInput{Key: "asset-freeze"})
	require.NoError(t, err)

	items, err := client.Midaz.Transactions.Balances.ListByAlias(ctx, "o", "l", alias)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	items, err = client.Midaz.Transactions.Balances.ListByAlias(ctx, "o", "l", "some-alias")
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFakeMidazBalancesListByExternalCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	externalCode := "some-code"
	account, err := client.Midaz.Onboarding.Accounts.Create(ctx, "o", "l", &midaz.CreateAccountInput{
		Name:         "External Lookup",
		Type:         "asset",
		AssetCode:    "BRL",
		ExternalCode: &externalCode,
	})
	require.NoError(t, err)

	created, err := client.Midaz.Transactions.Balances.CreateForAccount(ctx, "o", "l", account.ID, &midaz.CreateBalanceInput{Key: "asset-freeze"})
	require.NoError(t, err)

	items, err := client.Midaz.Transactions.Balances.ListByExternalCode(ctx, "o", "l", externalCode)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	items, err = client.Midaz.Transactions.Balances.ListByExternalCode(ctx, "o", "l", "missing-code")
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFakeMidazBalancesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("midaz.Balances.Create", injectedErr),
	)

	_, err := client.Midaz.Transactions.Balances.CreateForAccount(ctx, "o", "l", "acct-1", &midaz.CreateBalanceInput{Key: "asset-freeze"})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

// ===========================================================================
// 7. TransactionRoutes -- full CRUD
// ===========================================================================

func TestFakeMidazTransactionRoutesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	orgID := "org-1"
	ledgerID := "ledger-1"
	desc := "Transfer route"
	code := "XFER"

	// Create
	created, err := client.Midaz.Transactions.TransactionRoutes.Create(ctx, orgID, ledgerID, &midaz.CreateTransactionRouteInput{
		TransactionType: "transfer",
		Description:     &desc,
		Code:            &code,
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "transfer", created.TransactionType)
	assert.Equal(t, &desc, created.Description)
	assert.Equal(t, &code, created.Code)
	assert.Equal(t, orgID, created.OrganizationID)
	assert.Equal(t, ledgerID, created.LedgerID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Midaz.Transactions.TransactionRoutes.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "transfer", got.TransactionType)

	// Update
	newDesc := "Updated transfer route"
	updated, err := client.Midaz.Transactions.TransactionRoutes.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateTransactionRouteInput{
		Description: &newDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, &newDesc, updated.Description)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Transactions.TransactionRoutes.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Transactions.TransactionRoutes.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Transactions.TransactionRoutes.Get(ctx, orgID, ledgerID, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazTransactionRoutesNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-txroute-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error {
			_, err := client.Midaz.Transactions.TransactionRoutes.Get(ctx, "o", "l", ghost)
			return err
		}},
		{"Update", func() error {
			_, err := client.Midaz.Transactions.TransactionRoutes.Update(ctx, "o", "l", ghost, &midaz.UpdateTransactionRouteInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Transactions.TransactionRoutes.Delete(ctx, "o", "l", ghost) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

func TestFakeMidazTransactionRoutesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("midaz.TransactionRoutes.Create", injectedErr),
	)

	_, err := client.Midaz.Transactions.TransactionRoutes.Create(ctx, "o", "l", &midaz.CreateTransactionRouteInput{
		TransactionType: "transfer",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

// ===========================================================================
// 8. Operations -- query/update surface (Get, List, ListByTransaction, ListByAccount, Update)
// ===========================================================================

func TestFakeMidazOperationsGetNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	_, err := client.Midaz.Transactions.Operations.Get(ctx, "o", "l", "nonexistent-op-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazOperationsListEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	iter := client.Midaz.Transactions.Operations.List(ctx, "o", "l", nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFakeMidazOperationsListByTransactionEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	iter := client.Midaz.Transactions.Operations.ListByTransaction(ctx, "o", "l", "tx-999", nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFakeMidazOperationsListByAccountEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	iter := client.Midaz.Transactions.Operations.ListByAccount(ctx, "o", "l", "acct-999", nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFakeMidazOperationsErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("midaz.Operations.Get", injectedErr),
	)

	_, err := client.Midaz.Transactions.Operations.Get(ctx, "o", "l", "any-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

// ===========================================================================
// 9. OperationRoutes -- full CRUD
// ===========================================================================

func TestFakeMidazOperationRoutesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	orgID := "org-1"
	ledgerID := "ledger-1"
	desc := "Debit leg"

	// Create
	created, err := client.Midaz.Transactions.OperationRoutes.Create(ctx, orgID, ledgerID, &midaz.CreateOperationRouteInput{
		AccountID:   "acct-debit",
		Type:        "debit",
		Description: &desc,
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "acct-debit", created.AccountID)
	assert.Equal(t, "debit", created.Type)
	assert.Equal(t, &desc, created.Description)
	assert.Equal(t, orgID, created.OrganizationID)
	assert.Equal(t, ledgerID, created.LedgerID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Midaz.Transactions.OperationRoutes.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "debit", got.Type)

	// Update
	newAcctID := "acct-debit-v2"
	newAlias := "debit-alias"
	updated, err := client.Midaz.Transactions.OperationRoutes.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateOperationRouteInput{
		AccountID:    &newAcctID,
		AccountAlias: &newAlias,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "acct-debit-v2", updated.AccountID)
	assert.Equal(t, &newAlias, updated.AccountAlias)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Transactions.OperationRoutes.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Transactions.OperationRoutes.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Transactions.OperationRoutes.Get(ctx, orgID, ledgerID, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazOperationRoutesNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-oproute-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error {
			_, err := client.Midaz.Transactions.OperationRoutes.Get(ctx, "o", "l", ghost)
			return err
		}},
		{"Update", func() error {
			_, err := client.Midaz.Transactions.OperationRoutes.Update(ctx, "o", "l", ghost, &midaz.UpdateOperationRouteInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Transactions.OperationRoutes.Delete(ctx, "o", "l", ghost) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	}
}

func TestFakeMidazOperationRoutesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("midaz.OperationRoutes.Create", injectedErr),
	)

	_, err := client.Midaz.Transactions.OperationRoutes.Create(ctx, "o", "l", &midaz.CreateOperationRouteInput{
		AccountID: "acct-1",
		Type:      "debit",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

func TestFakeMidazOperationRoutesListMultiple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	for i := 0; i < 3; i++ {
		_, err := client.Midaz.Transactions.OperationRoutes.Create(ctx, "o", "l", &midaz.CreateOperationRouteInput{
			AccountID: "acct-x",
			Type:      "credit",
		})
		require.NoError(t, err)
	}

	iter := client.Midaz.Transactions.OperationRoutes.List(ctx, "o", "l", nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 3)
}
