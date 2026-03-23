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

func TestFakeFeesTransformTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	input := &fees.TransformTransactionInput{
		LedgerID: "ledger-123",
		Transaction: fees.TransactionDSL{
			Route:   "ted_out",
			Pending: true,
			Send: fees.TransactionDSLSend{
				Asset: "BRL",
				Value: "15000",
				Source: fees.TransactionDSLSource{
					From: []fees.TransactionDSLLeg{
						{AccountAlias: "@external/BRL", Share: &fees.TransactionDSLShare{Percentage: 100}},
					},
				},
				Distribute: fees.TransactionDSLDistribute{
					To: []fees.TransactionDSLLeg{
						{AccountAlias: "@customer", Share: &fees.TransactionDSLShare{Percentage: 100}},
					},
				},
			},
		},
	}

	output, err := client.Fees.Fees.TransformTransaction(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, input.Transaction, output.Transaction)
}

func TestFakeFeesTransformTransactionErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("fees.Fees.TransformTransaction", injectedErr),
	)

	_, err := client.Fees.Fees.TransformTransaction(ctx, &fees.TransformTransactionInput{
		LedgerID:    "ledger-123",
		Transaction: fees.TransactionDSL{},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

func TestFakeFeesTransformTransactionValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	output, err := client.Fees.Fees.TransformTransaction(ctx, nil)
	require.Error(t, err)
	assert.Nil(t, output)

	output, err = client.Fees.Fees.TransformTransaction(ctx, &fees.TransformTransactionInput{})
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "ledger ID is required")
}

func TestFakeFeesTransformTransactionReturnsDetachedCopy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	input := &fees.TransformTransactionInput{
		LedgerID: "ledger-123",
		Transaction: fees.TransactionDSL{
			Metadata: map[string]any{"packageAppliedID": "pkg-1"},
			Send: fees.TransactionDSLSend{
				Asset: "BRL",
				Value: "15000",
				Source: fees.TransactionDSLSource{
					From: []fees.TransactionDSLLeg{
						{Metadata: map[string]any{"feeLabel": "fee-a"}},
					},
				},
			},
		},
	}

	output, err := client.Fees.Fees.TransformTransaction(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	output.Transaction.Metadata["packageAppliedID"] = "pkg-2"
	output.Transaction.Send.Source.From[0].Metadata["feeLabel"] = "fee-b"

	assert.Equal(t, "pkg-1", input.Transaction.Metadata["packageAppliedID"])
	assert.Equal(t, "fee-a", input.Transaction.Send.Source.From[0].Metadata["feeLabel"])
}

func TestFakeFeesTransformTransactionReturnsDetachedCopyForDynamicFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	input := &fees.TransformTransactionInput{
		LedgerID: "ledger-123",
		Transaction: fees.TransactionDSL{
			TransactionDate: map[string]string{"at": "2026-01-01"},
			Send: fees.TransactionDSLSend{
				Asset: "BRL",
				Value: []string{"15000"},
				Distribute: fees.TransactionDSLDistribute{
					To: []fees.TransactionDSLLeg{
						{
							Amount: &fees.TransactionDSLAmount{
								Asset: "BRL",
								Value: []map[string]any{{"net": "14500"}},
							},
						},
					},
				},
			},
		},
	}

	output, err := client.Fees.Fees.TransformTransaction(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	outputDate := output.Transaction.TransactionDate.(map[string]string)
	outputDate["at"] = "changed"

	outputValue := output.Transaction.Send.Value.([]string)
	outputValue[0] = "changed"

	outputAmount := output.Transaction.Send.Distribute.To[0].Amount.Value.([]map[string]any)
	outputAmount[0]["net"] = "changed"

	inputDate := input.Transaction.TransactionDate.(map[string]string)
	inputValue := input.Transaction.Send.Value.([]string)
	inputAmount := input.Transaction.Send.Distribute.To[0].Amount.Value.([]map[string]any)

	assert.Equal(t, "2026-01-01", inputDate["at"])
	assert.Equal(t, "15000", inputValue[0])
	assert.Equal(t, "14500", inputAmount[0]["net"])
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
