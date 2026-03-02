package leriantest_test

import (
	"context"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/testing/leriantest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 1. Packages -- full CRUD
// ---------------------------------------------------------------------------

func TestFakeFeesPackagesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	flatAmount := int64(250)
	created, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		Name: "Standard Fees",
		Rules: []fees.FeeRule{
			{Type: "flat", Amount: &flatAmount, Currency: "BRL"},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Standard Fees", created.Name)
	assert.Equal(t, "active", created.Status)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Fees.Packages.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Standard Fees", got.Name)

	// Update
	updated, err := client.Fees.Packages.Update(ctx, created.ID, &fees.UpdatePackageInput{})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	iter := client.Fees.Packages.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)

	// Delete
	err = client.Fees.Packages.Delete(ctx, created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = client.Fees.Packages.Get(ctx, created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeFeesPackagesNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	ghost := "nonexistent-package-id"

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Get", func() error { _, err := client.Fees.Packages.Get(ctx, ghost); return err }},
		{"Update", func() error {
			_, err := client.Fees.Packages.Update(ctx, ghost, &fees.UpdatePackageInput{})
			return err
		}},
		{"Delete", func() error { return client.Fees.Packages.Delete(ctx, ghost) }},
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

// ---------------------------------------------------------------------------
// 2. Estimates -- RPC-style Calculate
// ---------------------------------------------------------------------------

func TestFakeFeesEstimatesCalculate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	estimate, err := client.Fees.Estimates.Calculate(ctx, &fees.CalculateEstimateInput{
		PackageID: "pkg-standard",
		Amount:    50000,
		Scale:     2,
		Currency:  "BRL",
	})
	require.NoError(t, err)
	require.NotNil(t, estimate)
	assert.NotEmpty(t, estimate.ID)

	// Calculate again -- each call produces a unique ID
	estimate2, err := client.Fees.Estimates.Calculate(ctx, &fees.CalculateEstimateInput{
		PackageID: "pkg-premium",
		Amount:    100000,
		Scale:     2,
		Currency:  "USD",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, estimate2.ID)
	assert.NotEqual(t, estimate.ID, estimate2.ID)
}

func TestFakeFeesEstimatesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("fees.Estimates.Calculate", injectedErr),
	)

	_, err := client.Fees.Estimates.Calculate(ctx, &fees.CalculateEstimateInput{
		PackageID: "pkg-1",
		Amount:    1000,
		Scale:     2,
		Currency:  "BRL",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

// ---------------------------------------------------------------------------
// 3. Fees -- RPC-style Calculate
// ---------------------------------------------------------------------------

func TestFakeFeesFeesCalculate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Calculate without a transaction ID
	fee, err := client.Fees.Fees.Calculate(ctx, &fees.CalculateFeeInput{
		PackageID: "pkg-standard",
		Amount:    75000,
		Scale:     2,
		Currency:  "BRL",
	})
	require.NoError(t, err)
	require.NotNil(t, fee)
	assert.NotEmpty(t, fee.ID)

	// Calculate with a transaction ID
	txID := "tx-abc-123"
	fee2, err := client.Fees.Fees.Calculate(ctx, &fees.CalculateFeeInput{
		PackageID:     "pkg-premium",
		TransactionID: &txID,
		Amount:        200000,
		Scale:         2,
		Currency:      "USD",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, fee2.ID)
	assert.NotEqual(t, fee.ID, fee2.ID)
}

func TestFakeFeesFeesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("fees.Fees.Calculate", injectedErr),
	)

	_, err := client.Fees.Fees.Calculate(ctx, &fees.CalculateFeeInput{
		PackageID: "pkg-1",
		Amount:    1000,
		Scale:     2,
		Currency:  "BRL",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

// ---------------------------------------------------------------------------
// Supplemental: List with multiple packages
// ---------------------------------------------------------------------------

func TestFakeFeesPackagesListMultipleItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create 3 packages
	for i := 0; i < 3; i++ {
		_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
			Name: "Package",
			Rules: []fees.FeeRule{
				{Type: "flat", Currency: "BRL"},
			},
		})
		require.NoError(t, err)
	}

	iter := client.Fees.Packages.List(ctx, nil)
	items, err := iter.Collect(ctx)
	require.NoError(t, err)
	assert.Len(t, items, 3)
}

// ---------------------------------------------------------------------------
// Supplemental: Packages error injection
// ---------------------------------------------------------------------------

func TestFakeFeesPackagesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("fees.Packages.Create", injectedErr),
	)

	_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		Name: "Should Fail",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)

	// Non-injected operations still work
	fee, err := client.Fees.Fees.Calculate(ctx, &fees.CalculateFeeInput{
		PackageID: "pkg-1",
		Amount:    1000,
		Scale:     2,
		Currency:  "BRL",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, fee.ID)
}

// ---------------------------------------------------------------------------
// Cross-service: all three fee services work together
// ---------------------------------------------------------------------------

func TestFakeFeesIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create a package
	flatAmount := int64(100)
	pkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		Name: "Integration Package",
		Rules: []fees.FeeRule{
			{Type: "flat", Amount: &flatAmount, Currency: "BRL"},
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, pkg.ID)

	// Use the package ID to calculate an estimate
	estimate, err := client.Fees.Estimates.Calculate(ctx, &fees.CalculateEstimateInput{
		PackageID: pkg.ID,
		Amount:    50000,
		Scale:     2,
		Currency:  "BRL",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, estimate.ID)

	// Use the package ID to calculate an actual fee
	fee, err := client.Fees.Fees.Calculate(ctx, &fees.CalculateFeeInput{
		PackageID: pkg.ID,
		Amount:    50000,
		Scale:     2,
		Currency:  "BRL",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, fee.ID)

	// Package is still retrievable
	gotPkg, err := client.Fees.Packages.Get(ctx, pkg.ID)
	require.NoError(t, err)
	assert.Equal(t, "Integration Package", gotPkg.Name)

	// Delete the package
	err = client.Fees.Packages.Delete(ctx, pkg.ID)
	require.NoError(t, err)

	// Package is gone
	_, err = client.Fees.Packages.Get(ctx, pkg.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
