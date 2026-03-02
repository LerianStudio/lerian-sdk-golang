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
	created, err := client.Midaz.AccountTypes.Create(ctx, orgID, ledgerID, &midaz.CreateAccountTypeInput{
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
	got, err := client.Midaz.AccountTypes.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "savings", got.Name)

	// Update
	newName := "deposit"
	updated, err := client.Midaz.AccountTypes.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateAccountTypeInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "deposit", updated.Name)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.AccountTypes.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.AccountTypes.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.AccountTypes.Get(ctx, orgID, ledgerID, created.ID)
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
		{"Get", func() error { _, err := client.Midaz.AccountTypes.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.AccountTypes.Update(ctx, "o", "l", ghost, &midaz.UpdateAccountTypeInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.AccountTypes.Delete(ctx, "o", "l", ghost) }},
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

	_, err := client.Midaz.AccountTypes.Create(ctx, "o", "l", &midaz.CreateAccountTypeInput{
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
	created, err := client.Midaz.Assets.Create(ctx, orgID, ledgerID, &midaz.CreateAssetInput{
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
	got, err := client.Midaz.Assets.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "BRL", got.Code)

	// Update
	newName := "Real Brasileiro"
	updated, err := client.Midaz.Assets.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateAssetInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Real Brasileiro", updated.Name)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Assets.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Assets.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Assets.Get(ctx, orgID, ledgerID, created.ID)
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
		{"Get", func() error { _, err := client.Midaz.Assets.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.Assets.Update(ctx, "o", "l", ghost, &midaz.UpdateAssetInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Assets.Delete(ctx, "o", "l", ghost) }},
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
		_, err := client.Midaz.Assets.Create(ctx, "o", "l", &midaz.CreateAssetInput{
			Name: "Asset",
			Code: "TST",
			Type: "currency",
		})
		require.NoError(t, err)
	}

	iter := client.Midaz.Assets.List(ctx, "o", "l", nil)
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
	created, err := client.Midaz.AssetRates.Create(ctx, orgID, ledgerID, &midaz.CreateAssetRateInput{
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
	got, err := client.Midaz.AssetRates.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "BRL", got.BaseAssetCode)

	// Update
	newAmount := int64(560)
	newScale := 2
	updated, err := client.Midaz.AssetRates.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateAssetRateInput{
		Amount: &newAmount,
		Scale:  &newScale,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, int64(560), updated.Amount)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.AssetRates.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.AssetRates.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.AssetRates.Get(ctx, orgID, ledgerID, created.ID)
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
		{"Get", func() error { _, err := client.Midaz.AssetRates.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.AssetRates.Update(ctx, "o", "l", ghost, &midaz.UpdateAssetRateInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.AssetRates.Delete(ctx, "o", "l", ghost) }},
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

	created, err := client.Midaz.AssetRates.Create(ctx, "o", "l", &midaz.CreateAssetRateInput{
		BaseAssetCode:    "BRL",
		CounterAssetCode: "EUR",
		Amount:           600,
		Scale:            2,
		ExternalID:       &externalID,
	})
	require.NoError(t, err)

	// Found
	got, err := client.Midaz.AssetRates.GetByExternalID(ctx, "o", "l", "ext-rate-lookup")
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Not found
	_, err = client.Midaz.AssetRates.GetByExternalID(ctx, "o", "l", "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazAssetRatesGetFromAssetCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	created, err := client.Midaz.AssetRates.Create(ctx, "o", "l", &midaz.CreateAssetRateInput{
		BaseAssetCode:    "GBP",
		CounterAssetCode: "USD",
		Amount:           130,
		Scale:            2,
	})
	require.NoError(t, err)

	// Found by base asset code
	got, err := client.Midaz.AssetRates.GetFromAssetCode(ctx, "o", "l", "GBP")
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Not found
	_, err = client.Midaz.AssetRates.GetFromAssetCode(ctx, "o", "l", "JPY")
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
	created, err := client.Midaz.Portfolios.Create(ctx, orgID, ledgerID, &midaz.CreatePortfolioInput{
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
	got, err := client.Midaz.Portfolios.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Main Portfolio", got.Name)

	// Update
	newName := "Updated Portfolio"
	updated, err := client.Midaz.Portfolios.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdatePortfolioInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Updated Portfolio", updated.Name)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Portfolios.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Portfolios.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Portfolios.Get(ctx, orgID, ledgerID, created.ID)
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
		{"Get", func() error { _, err := client.Midaz.Portfolios.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.Portfolios.Update(ctx, "o", "l", ghost, &midaz.UpdatePortfolioInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Portfolios.Delete(ctx, "o", "l", ghost) }},
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

	_, err := client.Midaz.Portfolios.Create(ctx, "o", "l", &midaz.CreatePortfolioInput{
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
	created, err := client.Midaz.Segments.Create(ctx, orgID, ledgerID, &midaz.CreateSegmentInput{
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
	got, err := client.Midaz.Segments.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Retail Banking", got.Name)

	// Update
	newName := "Corporate Banking"
	updated, err := client.Midaz.Segments.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateSegmentInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Corporate Banking", updated.Name)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Segments.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Segments.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Segments.Get(ctx, orgID, ledgerID, created.ID)
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
		{"Get", func() error { _, err := client.Midaz.Segments.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.Segments.Update(ctx, "o", "l", ghost, &midaz.UpdateSegmentInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Segments.Delete(ctx, "o", "l", ghost) }},
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
		_, err := client.Midaz.Segments.Create(ctx, "o", "l", &midaz.CreateSegmentInput{
			Name: "Segment",
		})
		require.NoError(t, err)
	}

	iter := client.Midaz.Segments.List(ctx, "o", "l", nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 3)
}

// ===========================================================================
// 6. Balances -- full CRUD + GetByAlias + GetByExternalCode + GetByAccountID
// ===========================================================================

func TestFakeMidazBalancesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	orgID := "org-1"
	ledgerID := "ledger-1"

	// Create
	created, err := client.Midaz.Balances.Create(ctx, orgID, ledgerID, &midaz.CreateBalanceInput{
		AccountID: "acct-123",
		AssetCode: "BRL",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "acct-123", created.AccountID)
	assert.Equal(t, "BRL", created.AssetCode)
	assert.Equal(t, "active", created.Status.Code)
	assert.True(t, created.AllowSending)
	assert.True(t, created.AllowReceiving)
	assert.Equal(t, orgID, created.OrganizationID)
	assert.Equal(t, ledgerID, created.LedgerID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Midaz.Balances.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "acct-123", got.AccountID)

	// Update
	allowSending := false
	updated, err := client.Midaz.Balances.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateBalanceInput{
		AllowSending: &allowSending,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.False(t, updated.AllowSending)
	assert.True(t, updated.AllowReceiving) // unchanged
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.Balances.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.Balances.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.Balances.Get(ctx, orgID, ledgerID, created.ID)
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
		{"Get", func() error { _, err := client.Midaz.Balances.Get(ctx, "o", "l", ghost); return err }},
		{"Update", func() error {
			_, err := client.Midaz.Balances.Update(ctx, "o", "l", ghost, &midaz.UpdateBalanceInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.Balances.Delete(ctx, "o", "l", ghost) }},
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

func TestFakeMidazBalancesGetByAccountID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	created, err := client.Midaz.Balances.Create(ctx, "o", "l", &midaz.CreateBalanceInput{
		AccountID: "acct-lookup-001",
		AssetCode: "USD",
	})
	require.NoError(t, err)

	// Found
	got, err := client.Midaz.Balances.GetByAccountID(ctx, "o", "l", "acct-lookup-001")
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Not found
	_, err = client.Midaz.Balances.GetByAccountID(ctx, "o", "l", "nonexistent-acct")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazBalancesGetByAlias(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// The Balance model has AccountAlias but CreateBalanceInput does not set it
	// directly. GetByAlias searches by AccountAlias field on the stored Balance.
	// Since the fake Create does not set AccountAlias, this should return not found.
	_, err := client.Midaz.Balances.GetByAlias(ctx, "o", "l", "some-alias")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazBalancesGetByExternalCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// GetByExternalCode on Balance always returns not found in the fake
	// because the Balance model does not have an ExternalCode field.
	_, err := client.Midaz.Balances.GetByExternalCode(ctx, "o", "l", "some-code")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazBalancesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("midaz.Balances.Create", injectedErr),
	)

	_, err := client.Midaz.Balances.Create(ctx, "o", "l", &midaz.CreateBalanceInput{
		AccountID: "acct-1",
		AssetCode: "BRL",
	})
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
	created, err := client.Midaz.TransactionRoutes.Create(ctx, orgID, ledgerID, &midaz.CreateTransactionRouteInput{
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
	got, err := client.Midaz.TransactionRoutes.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "transfer", got.TransactionType)

	// Update
	newDesc := "Updated transfer route"
	updated, err := client.Midaz.TransactionRoutes.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateTransactionRouteInput{
		Description: &newDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, &newDesc, updated.Description)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.TransactionRoutes.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.TransactionRoutes.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.TransactionRoutes.Get(ctx, orgID, ledgerID, created.ID)
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
			_, err := client.Midaz.TransactionRoutes.Get(ctx, "o", "l", ghost)
			return err
		}},
		{"Update", func() error {
			_, err := client.Midaz.TransactionRoutes.Update(ctx, "o", "l", ghost, &midaz.UpdateTransactionRouteInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.TransactionRoutes.Delete(ctx, "o", "l", ghost) }},
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

	_, err := client.Midaz.TransactionRoutes.Create(ctx, "o", "l", &midaz.CreateTransactionRouteInput{
		TransactionType: "transfer",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

// ===========================================================================
// 8. Operations -- read-only (Get, List, ListByTransaction, ListByAccount)
// ===========================================================================

func TestFakeMidazOperationsGetNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	_, err := client.Midaz.Operations.Get(ctx, "o", "l", "nonexistent-op-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeMidazOperationsListEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	iter := client.Midaz.Operations.List(ctx, "o", "l", nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFakeMidazOperationsListByTransactionEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	iter := client.Midaz.Operations.ListByTransaction(ctx, "o", "l", "tx-999", nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFakeMidazOperationsListByAccountEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	iter := client.Midaz.Operations.ListByAccount(ctx, "o", "l", "acct-999", nil)
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

	_, err := client.Midaz.Operations.Get(ctx, "o", "l", "any-id")
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
	created, err := client.Midaz.OperationRoutes.Create(ctx, orgID, ledgerID, &midaz.CreateOperationRouteInput{
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
	got, err := client.Midaz.OperationRoutes.Get(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "debit", got.Type)

	// Update
	newAcctID := "acct-debit-v2"
	newAlias := "debit-alias"
	updated, err := client.Midaz.OperationRoutes.Update(ctx, orgID, ledgerID, created.ID, &midaz.UpdateOperationRouteInput{
		AccountID:    &newAcctID,
		AccountAlias: &newAlias,
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "acct-debit-v2", updated.AccountID)
	assert.Equal(t, &newAlias, updated.AccountAlias)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Midaz.OperationRoutes.List(ctx, orgID, ledgerID, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Midaz.OperationRoutes.Delete(ctx, orgID, ledgerID, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Midaz.OperationRoutes.Get(ctx, orgID, ledgerID, created.ID)
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
			_, err := client.Midaz.OperationRoutes.Get(ctx, "o", "l", ghost)
			return err
		}},
		{"Update", func() error {
			_, err := client.Midaz.OperationRoutes.Update(ctx, "o", "l", ghost, &midaz.UpdateOperationRouteInput{})
			return err
		}},
		{"Delete", func() error { return client.Midaz.OperationRoutes.Delete(ctx, "o", "l", ghost) }},
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

	_, err := client.Midaz.OperationRoutes.Create(ctx, "o", "l", &midaz.CreateOperationRouteInput{
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
		_, err := client.Midaz.OperationRoutes.Create(ctx, "o", "l", &midaz.CreateOperationRouteInput{
			AccountID: "acct-x",
			Type:      "credit",
		})
		require.NoError(t, err)
	}

	iter := client.Midaz.OperationRoutes.List(ctx, "o", "l", nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 3)
}
