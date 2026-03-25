package leriantest_test

import (
	"context"
	"testing"
	"time"

	"github.com/LerianStudio/lerian-sdk-golang/fees"
	"github.com/LerianStudio/lerian-sdk-golang/testing/leriantest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func boolPtr(v bool) *bool       { return &v }
func stringPtr(v string) *string { return &v }

func sampleFeeMap() map[string]fees.Fee {
	return map[string]fees.Fee{
		"transfer_fee": {
			FeeLabel: "Transfer Fee",
			CalculationModel: &fees.CalculationModel{
				ApplicationRule: "flatFee",
				Calculations: []fees.Calculation{
					{Type: "flat", Value: "2.50"},
				},
			},
			ReferenceAmount:  "originalAmount",
			Priority:         1,
			IsDeductibleFrom: boolPtr(false),
			CreditAccount:    "@fees/BRL",
		},
	}
}

func sampleTransactionDSL() fees.TransactionDSL {
	return fees.TransactionDSL{
		Route:   "ted_out",
		Pending: true,
		Send: fees.TransactionDSLSend{
			Asset: "BRL",
			Value: "150.00",
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
	}
}

// ---------------------------------------------------------------------------
// 1. Packages -- full CRUD
// ---------------------------------------------------------------------------

func TestFakeFeesPackagesCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create
	created, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Standard Fees",
		LedgerID:      "ledger-001",
		MinimumAmount: "100",
		MaximumAmount: "1000",
		Fees:          sampleFeeMap(),
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, "Standard Fees", created.FeeGroupLabel)
	assert.Equal(t, "ledger-001", created.LedgerID)
	assert.Equal(t, "100", created.MinimumAmount)
	assert.Equal(t, "1000", created.MaximumAmount)
	assert.NotNil(t, created.Enable)
	assert.True(t, *created.Enable)
	assert.Len(t, created.Fees, 1)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	// Get
	got, err := client.Fees.Packages.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Standard Fees", got.FeeGroupLabel)

	// Update
	updated, err := client.Fees.Packages.Update(ctx, created.ID, &fees.UpdatePackageInput{
		FeeGroupLabel: "Premium Fees",
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Premium Fees", updated.FeeGroupLabel)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))

	// List -- should have 1 item
	resp, err := client.Fees.Packages.List(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, 1, resp.TotalItems)
	assert.Equal(t, created.ID, resp.Items[0].ID)

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

func TestFakeFeesPackagesCreateAndUpdateNilInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	created, err := client.Fees.Packages.Create(ctx, nil)
	require.Error(t, err)
	assert.Nil(t, created)
	assert.Contains(t, err.Error(), "input is required")

	existing, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Standard Fees",
		LedgerID:      "ledger-001",
		MinimumAmount: "100.00",
		MaximumAmount: "1000.00",
		Fees:          sampleFeeMap(),
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)

	updated, err := client.Fees.Packages.Update(ctx, existing.ID, nil)
	require.Error(t, err)
	assert.Nil(t, updated)
	assert.Contains(t, err.Error(), "input is required")
}

// ---------------------------------------------------------------------------
// 2. Estimates -- RPC-style Calculate
// ---------------------------------------------------------------------------

func TestFakeFeesEstimatesCalculate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	pkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel:    "Standard Fees",
		TransactionRoute: stringPtr("ted_out"),
		LedgerID:         "ledger-001",
		MinimumAmount:    "100.00",
		MaximumAmount:    "1000.00",
		Fees:             sampleFeeMap(),
		Enable:           boolPtr(true),
	})
	require.NoError(t, err)

	resp, err := client.Fees.Estimates.Calculate(ctx, &fees.FeeEstimateInput{
		PackageID:   pkg.ID,
		LedgerID:    "ledger-001",
		Transaction: sampleTransactionDSL(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Successfully estimated fee.", resp.Message)
	assert.NotNil(t, resp.FeesApplied)
	assert.Equal(t, "ledger-001", resp.FeesApplied.LedgerID)
	assert.Equal(t, "152.50", resp.FeesApplied.Transaction.Send.Value)
	assert.Equal(t, pkg.ID, resp.FeesApplied.Transaction.Metadata["packageAppliedID"])
	require.Len(t, resp.FeesApplied.Transaction.Send.Distribute.To, 2)
	assert.Equal(t, "@fees/BRL", resp.FeesApplied.Transaction.Send.Distribute.To[1].AccountAlias)
	assert.Equal(t, "Transfer Fee", resp.FeesApplied.Transaction.Send.Distribute.To[1].Metadata["feeLabel"])

	// Calculate again -- each call returns a valid response
	pkg2, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Premium Fees",
		LedgerID:      "ledger-002",
		MinimumAmount: "100.00",
		MaximumAmount: "1000.00",
		Fees:          sampleFeeMap(),
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)

	resp2, err := client.Fees.Estimates.Calculate(ctx, &fees.FeeEstimateInput{
		PackageID:   pkg2.ID,
		LedgerID:    "ledger-002",
		Transaction: sampleTransactionDSL(),
	})
	require.NoError(t, err)
	assert.NotNil(t, resp2.FeesApplied)
	assert.Equal(t, "ledger-002", resp2.FeesApplied.LedgerID)
}

func TestFakeFeesEstimatesCalculateNoFeesApplied(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	pkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "No Fees",
		LedgerID:      "ledger-001",
		MinimumAmount: "100.00",
		MaximumAmount: "1000.00",
		Fees:          map[string]fees.Fee{},
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)

	resp, err := client.Fees.Estimates.Calculate(ctx, &fees.FeeEstimateInput{
		PackageID:   pkg.ID,
		LedgerID:    "ledger-001",
		Transaction: sampleTransactionDSL(),
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.FeesApplied)
	assert.Equal(t, "No fee or gratuity rules were found for the given parameters.", resp.Message)
}

func TestFakeFeesEstimatesCalculatePackageNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	resp, err := client.Fees.Estimates.Calculate(ctx, &fees.FeeEstimateInput{
		PackageID:   "pkg-missing",
		LedgerID:    "ledger-001",
		Transaction: sampleTransactionDSL(),
	})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "not found")
}

func TestFakeFeesEstimatesCalculateRespectsPackageScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *fees.CreatePackageInput
		estimate *fees.FeeEstimateInput
	}{
		{
			name: "disabled package",
			input: &fees.CreatePackageInput{
				FeeGroupLabel: "Disabled Fees",
				LedgerID:      "ledger-001",
				MinimumAmount: "100.00",
				MaximumAmount: "1000.00",
				Fees:          sampleFeeMap(),
				Enable:        boolPtr(false),
			},
			estimate: &fees.FeeEstimateInput{LedgerID: "ledger-001", Transaction: sampleTransactionDSL()},
		},
		{
			name: "wrong ledger",
			input: &fees.CreatePackageInput{
				FeeGroupLabel: "Other Ledger",
				LedgerID:      "ledger-002",
				MinimumAmount: "100.00",
				MaximumAmount: "1000.00",
				Fees:          sampleFeeMap(),
				Enable:        boolPtr(true),
			},
			estimate: &fees.FeeEstimateInput{LedgerID: "ledger-001", Transaction: sampleTransactionDSL()},
		},
		{
			name: "wrong route",
			input: &fees.CreatePackageInput{
				FeeGroupLabel:    "Wrong Route",
				LedgerID:         "ledger-001",
				TransactionRoute: stringPtr("pix_out"),
				MinimumAmount:    "100.00",
				MaximumAmount:    "1000.00",
				Fees:             sampleFeeMap(),
				Enable:           boolPtr(true),
			},
			estimate: &fees.FeeEstimateInput{LedgerID: "ledger-001", Transaction: sampleTransactionDSL()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			client := leriantest.NewFakeClient()

			pkg, err := client.Fees.Packages.Create(ctx, tt.input)
			require.NoError(t, err)

			estimate := *tt.estimate
			estimate.PackageID = pkg.ID
			resp, err := client.Fees.Estimates.Calculate(ctx, &estimate)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Nil(t, resp.FeesApplied)
			assert.Equal(t, "No fee or gratuity rules were found for the given parameters.", resp.Message)
		})
	}
}

func TestFakeFeesEstimatesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("fees.Estimates.Calculate", injectedErr),
	)

	_, err := client.Fees.Estimates.Calculate(ctx, &fees.FeeEstimateInput{
		PackageID:   "pkg-1",
		LedgerID:    "ledger-001",
		Transaction: sampleTransactionDSL(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

func TestFakeFeesEstimatesNilInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	resp, err := client.Fees.Estimates.Calculate(ctx, nil)
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestFakeFeesEstimatesValidationParity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   *fees.FeeEstimateInput
		message string
	}{
		{name: "missing asset", input: &fees.FeeEstimateInput{PackageID: "pkg-1", LedgerID: "ledger-1", Transaction: fees.TransactionDSL{Send: fees.TransactionDSLSend{Value: "10.00", Source: fees.TransactionDSLSource{From: []fees.TransactionDSLLeg{{AccountAlias: "sender", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}, Distribute: fees.TransactionDSLDistribute{To: []fees.TransactionDSLLeg{{AccountAlias: "recipient", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}}}}, message: "transaction send asset is required"},
		{name: "missing value", input: &fees.FeeEstimateInput{PackageID: "pkg-1", LedgerID: "ledger-1", Transaction: fees.TransactionDSL{Send: fees.TransactionDSLSend{Asset: "BRL", Source: fees.TransactionDSLSource{From: []fees.TransactionDSLLeg{{AccountAlias: "sender", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}, Distribute: fees.TransactionDSLDistribute{To: []fees.TransactionDSLLeg{{AccountAlias: "recipient", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}}}}, message: "transaction send value is required"},
		{name: "missing source legs", input: &fees.FeeEstimateInput{PackageID: "pkg-1", LedgerID: "ledger-1", Transaction: fees.TransactionDSL{Send: fees.TransactionDSLSend{Asset: "BRL", Value: "10.00", Distribute: fees.TransactionDSLDistribute{To: []fees.TransactionDSLLeg{{AccountAlias: "recipient", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}}}}, message: "transaction source legs are required"},
		{name: "missing distribute legs", input: &fees.FeeEstimateInput{PackageID: "pkg-1", LedgerID: "ledger-1", Transaction: fees.TransactionDSL{Send: fees.TransactionDSLSend{Asset: "BRL", Value: "10.00", Source: fees.TransactionDSLSource{From: []fees.TransactionDSLLeg{{AccountAlias: "sender", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}}}}, message: "transaction distribute legs are required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			client := leriantest.NewFakeClient()

			resp, err := client.Fees.Estimates.Calculate(ctx, tt.input)
			require.Error(t, err)
			assert.Nil(t, resp)
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Fees -- RPC-style Calculate
// ---------------------------------------------------------------------------

func TestFakeFeesFeesCalculate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel:    "Disabled Fees",
		TransactionRoute: stringPtr("ted_out"),
		LedgerID:         "ledger-001",
		MinimumAmount:    "100.00",
		MaximumAmount:    "1000.00",
		Fees:             sampleFeeMap(),
		Enable:           boolPtr(false),
	})
	require.NoError(t, err)
	matched, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel:    "Matched Fees",
		SegmentID:        stringPtr("seg-retail"),
		TransactionRoute: stringPtr("ted_out"),
		LedgerID:         "ledger-002",
		MinimumAmount:    "100.00",
		MaximumAmount:    "1000.00",
		Fees:             sampleFeeMap(),
		Enable:           boolPtr(true),
	})
	require.NoError(t, err)

	result, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: sampleTransactionDSL(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "ledger-001", result.LedgerID)
	assert.Equal(t, "150.00", result.Transaction.Send.Value)
	assert.Nil(t, result.Transaction.Metadata)

	// Calculate with segment
	result2, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		SegmentID:   stringPtr("seg-retail"),
		LedgerID:    "ledger-002",
		Transaction: sampleTransactionDSL(),
	})
	require.NoError(t, err)
	assert.Equal(t, "ledger-002", result2.LedgerID)
	assert.NotNil(t, result2.SegmentID)
	assert.Equal(t, "seg-retail", *result2.SegmentID)
	assert.Equal(t, matched.ID, result2.Transaction.Metadata["packageAppliedID"])
	assert.Equal(t, "152.50", result2.Transaction.Send.Value)
	require.Len(t, result2.Transaction.Send.Source.From, 2)
	require.Len(t, result2.Transaction.Send.Distribute.To, 2)
	assert.Equal(t, "@fees/BRL", result2.Transaction.Send.Distribute.To[1].AccountAlias)
	assert.Equal(t, "Transfer Fee", result2.Transaction.Send.Distribute.To[1].Metadata["feeLabel"])
}

func TestFakeFeesPackagesUpdatePreservesFeesWhenFeesPatchOmitted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	created, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Original",
		LedgerID:      "ledger-001",
		MinimumAmount: "100.00",
		MaximumAmount: "1000.00",
		Fees:          sampleFeeMap(),
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)
	require.Len(t, created.Fees, 1)

	updated, err := client.Fees.Packages.Update(ctx, created.ID, &fees.UpdatePackageInput{FeeGroupLabel: "Renamed"})
	require.NoError(t, err)
	require.Len(t, updated.Fees, 1)
	assert.Contains(t, updated.Fees, "transfer_fee")
}

func TestFakeFeesFeesCalculatePreservesOriginalLegFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Standard Fees",
		LedgerID:      "ledger-001",
		MinimumAmount: "100.00",
		MaximumAmount: "1000.00",
		Fees:          sampleFeeMap(),
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)

	input := &fees.FeeCalculate{
		LedgerID: "ledger-001",
		Transaction: fees.TransactionDSL{Send: fees.TransactionDSLSend{
			Asset: "BRL",
			Value: "150.00",
			Source: fees.TransactionDSLSource{From: []fees.TransactionDSLLeg{{
				AccountAlias: "@external/BRL",
				BalanceKey:   "available",
				Description:  "original source leg",
				Metadata:     map[string]any{"origin": "source"},
				Amount:       &fees.TransactionDSLAmount{Asset: "BRL", Value: "150.00"},
			}}},
			Distribute: fees.TransactionDSLDistribute{To: []fees.TransactionDSLLeg{{
				AccountAlias: "@merchant/BRL",
				Description:  "recipient leg",
				Metadata:     map[string]any{"origin": "recipient"},
				Amount:       &fees.TransactionDSLAmount{Asset: "BRL", Value: "150.00"},
			}}},
		}},
	}

	result, err := client.Fees.Fees.Calculate(ctx, input)
	require.NoError(t, err)
	require.Len(t, result.Transaction.Send.Source.From, 2)
	assert.Equal(t, "available", result.Transaction.Send.Source.From[0].BalanceKey)
	assert.Equal(t, "original source leg", result.Transaction.Send.Source.From[0].Description)
	assert.Equal(t, "source", result.Transaction.Send.Source.From[0].Metadata["origin"])
}

func TestFakeFeesFeesCalculateRejectsAmbiguousPackages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	for i := 0; i < 2; i++ {
		_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
			FeeGroupLabel:    "Ambiguous",
			TransactionRoute: stringPtr("ted_out"),
			LedgerID:         "ledger-001",
			MinimumAmount:    "100.00",
			MaximumAmount:    "1000.00",
			Fees:             sampleFeeMap(),
			Enable:           boolPtr(true),
		})
		require.NoError(t, err)
	}

	result, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: sampleTransactionDSL(),
	})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "more than one package matched")
}

func TestFakeFeesFeesCalculateMatchesGlobalPackagesWhenRouteAndSegmentAreProvided(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	pkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Global Fees",
		LedgerID:      "ledger-001",
		MinimumAmount: "100.00",
		MaximumAmount: "1000.00",
		Fees:          sampleFeeMap(),
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)

	segmentID := "seg-retail"
	result, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		SegmentID: &segmentID,
		LedgerID:  "ledger-001",
		Transaction: fees.TransactionDSL{
			Route: "wire_transfer",
			Send:  sampleTransactionDSL().Send,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, pkg.ID, result.Transaction.Metadata["packageAppliedID"])
}

func TestFakeFeesFeesCalculatePrefersMoreSpecificPackage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	route := "ted_out"
	segmentID := "seg-retail"

	globalPkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Global Fees",
		LedgerID:      "ledger-001",
		MinimumAmount: "100.00",
		MaximumAmount: "1000.00",
		Fees:          sampleFeeMap(),
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)

	specificPkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel:    "Specific Fees",
		LedgerID:         "ledger-001",
		SegmentID:        &segmentID,
		TransactionRoute: &route,
		MinimumAmount:    "100.00",
		MaximumAmount:    "1000.00",
		Fees: map[string]fees.Fee{
			"transfer_fee": {
				FeeLabel: "Transfer Fee",
				CalculationModel: &fees.CalculationModel{
					ApplicationRule: "flatFee",
					Calculations:    []fees.Calculation{{Type: "flat", Value: "5.00"}},
				},
				ReferenceAmount:  "originalAmount",
				Priority:         1,
				IsDeductibleFrom: boolPtr(false),
				CreditAccount:    "@fees/BRL",
			},
		},
		Enable: boolPtr(true),
	})
	require.NoError(t, err)

	result, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		SegmentID:   &segmentID,
		LedgerID:    "ledger-001",
		Transaction: sampleTransactionDSL(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEqual(t, globalPkg.ID, result.Transaction.Metadata["packageAppliedID"])
	assert.Equal(t, specificPkg.ID, result.Transaction.Metadata["packageAppliedID"])
	assert.Equal(t, "155.00", result.Transaction.Send.Value)
}

func TestFakeFeesFeesCalculateDoesNotMatchOutOfRangeSpecificPackage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	route := "ted_out"

	_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel:    "Specific Out Of Range",
		LedgerID:         "ledger-001",
		TransactionRoute: &route,
		MinimumAmount:    "1000.01",
		MaximumAmount:    "2000.00",
		Fees:             sampleFeeMap(),
		Enable:           boolPtr(true),
	})
	require.NoError(t, err)

	result, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: sampleTransactionDSL(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Transaction.Metadata)
	assert.Equal(t, "150.00", result.Transaction.Send.Value)
}

func TestFakeFeesFeesCalculatePreservesDuplicateLegs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Duplicate Legs",
		LedgerID:      "ledger-001",
		MinimumAmount: "100.00",
		MaximumAmount: "1000.00",
		Fees: map[string]fees.Fee{
			"transfer_fee": {
				FeeLabel: "Transfer Fee",
				CalculationModel: &fees.CalculationModel{
					ApplicationRule: "flatFee",
					Calculations:    []fees.Calculation{{Type: "flat", Value: "2.50"}},
				},
				ReferenceAmount:  "originalAmount",
				Priority:         1,
				IsDeductibleFrom: boolPtr(false),
				CreditAccount:    "@fees/BRL",
			},
		},
		Enable: boolPtr(true),
	})
	require.NoError(t, err)

	result, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		LedgerID: "ledger-001",
		Transaction: fees.TransactionDSL{
			Send: fees.TransactionDSLSend{
				Asset: "BRL",
				Value: "150.00",
				Source: fees.TransactionDSLSource{From: []fees.TransactionDSLLeg{
					{AccountAlias: "@sender", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "100.00"}},
					{AccountAlias: "@sender", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "50.00"}},
				}},
				Distribute: fees.TransactionDSLDistribute{To: []fees.TransactionDSLLeg{{AccountAlias: "@receiver", Share: &fees.TransactionDSLShare{Percentage: 100}}}},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "152.50", result.Transaction.Send.Value)

	baseValues := make([]any, 0, 2)
	for _, leg := range result.Transaction.Send.Source.From {
		if leg.AccountAlias == "@sender" && leg.Route == "" && leg.Amount != nil {
			baseValues = append(baseValues, leg.Amount.Value)
		}
	}
	assert.Contains(t, baseValues, "100.00")
	assert.Contains(t, baseValues, "50.00")
}

func TestFakeFeesFeesCalculateHonorsRoutesAndWaivedAccounts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	waived := []string{"@external/BRL->primary"}
	routeFrom := "fee_debit"
	routeTo := "fee_credit"
	_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel:    "Route Aware",
		TransactionRoute: stringPtr("ted_out"),
		LedgerID:         "ledger-777",
		MinimumAmount:    "100.00",
		MaximumAmount:    "1000.00",
		WaivedAccounts:   &waived,
		Fees: map[string]fees.Fee{
			"transfer_fee": {
				FeeLabel: "Transfer Fee",
				CalculationModel: &fees.CalculationModel{
					ApplicationRule: "flatFee",
					Calculations:    []fees.Calculation{{Type: "flat", Value: "10.00"}},
				},
				ReferenceAmount:  "originalAmount",
				Priority:         1,
				IsDeductibleFrom: boolPtr(false),
				CreditAccount:    "@fees/BRL",
				RouteFrom:        &routeFrom,
				RouteTo:          &routeTo,
			},
		},
		Enable: boolPtr(true),
	})
	require.NoError(t, err)

	result, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		LedgerID: "ledger-777",
		Transaction: fees.TransactionDSL{
			Route: "ted_out",
			Send: fees.TransactionDSLSend{
				Asset: "BRL",
				Value: "200.00",
				Source: fees.TransactionDSLSource{From: []fees.TransactionDSLLeg{
					{AccountAlias: "@external/BRL", Route: "primary", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "100.00"}},
					{AccountAlias: "@partner/BRL", Route: "secondary", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "100.00"}},
				}},
				Distribute: fees.TransactionDSLDistribute{To: []fees.TransactionDSLLeg{{AccountAlias: "@customer", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "200.00"}}}},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "210.00", result.Transaction.Send.Value)

	var waivedOriginal *fees.TransactionDSLLeg
	var chargedFeeLeg *fees.TransactionDSLLeg
	for i := range result.Transaction.Send.Source.From {
		leg := &result.Transaction.Send.Source.From[i]
		switch {
		case leg.AccountAlias == "@external/BRL" && leg.Route == "primary":
			waivedOriginal = leg
		case leg.AccountAlias == "@partner/BRL" && leg.Route == "fee_debit":
			chargedFeeLeg = leg
		}
	}
	require.NotNil(t, waivedOriginal)
	require.NotNil(t, chargedFeeLeg)
	assert.Equal(t, "100.00", waivedOriginal.Amount.Value)
	assert.Equal(t, "10.00", chargedFeeLeg.Amount.Value)

	var routeToLeg *fees.TransactionDSLLeg
	for i := range result.Transaction.Send.Distribute.To {
		leg := &result.Transaction.Send.Distribute.To[i]
		if leg.AccountAlias == "@fees/BRL" {
			routeToLeg = leg
			break
		}
	}
	require.NotNil(t, routeToLeg)
	assert.Equal(t, "fee_credit", routeToLeg.Route)
	assert.Equal(t, "Transfer Fee", routeToLeg.Metadata["feeLabel"])
}

func TestFakeFeesFeesCalculateUsesAfterFeesAmountAfterDeductibleFees(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Mixed Fees",
		LedgerID:      "ledger-001",
		MinimumAmount: "0.00",
		MaximumAmount: "1000.00",
		Fees: map[string]fees.Fee{
			"deductible_flat": {
				FeeLabel:         "Deductible Flat",
				CalculationModel: &fees.CalculationModel{ApplicationRule: "flatFee", Calculations: []fees.Calculation{{Type: "flat", Value: "10.00"}}},
				ReferenceAmount:  "originalAmount",
				Priority:         1,
				IsDeductibleFrom: boolPtr(true),
				CreditAccount:    "@fees/BRL",
			},
			"after_fee_percent": {
				FeeLabel:         "After Fee Percent",
				CalculationModel: &fees.CalculationModel{ApplicationRule: "percentual", Calculations: []fees.Calculation{{Type: "percentage", Value: "10.00"}}},
				ReferenceAmount:  "afterFeesAmount",
				Priority:         2,
				IsDeductibleFrom: boolPtr(false),
				CreditAccount:    "@fees/BRL",
			},
		},
		Enable: boolPtr(true),
	})
	require.NoError(t, err)

	result, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		LedgerID: "ledger-001",
		Transaction: fees.TransactionDSL{Send: fees.TransactionDSLSend{
			Asset:      "BRL",
			Value:      "100.00",
			Source:     fees.TransactionDSLSource{From: []fees.TransactionDSLLeg{{AccountAlias: "@sender/BRL", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "100.00"}}}},
			Distribute: fees.TransactionDSLDistribute{To: []fees.TransactionDSLLeg{{AccountAlias: "@recipient/BRL", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "100.00"}}}},
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, "109.00", result.Transaction.Send.Value)
	require.Len(t, result.Transaction.Send.Distribute.To, 3)

	amountsByLabel := map[string]string{}
	recipientAmount := ""
	for _, leg := range result.Transaction.Send.Distribute.To {
		if leg.AccountAlias == "@recipient/BRL" {
			recipientAmount = leg.Amount.Value.(string)
			continue
		}
		if label, ok := leg.Metadata["feeLabel"].(string); ok {
			amountsByLabel[label] = leg.Amount.Value.(string)
		}
	}

	assert.Equal(t, "90.00", recipientAmount)
	assert.Equal(t, "10.00", amountsByLabel["Deductible Flat"])
	assert.Equal(t, "9.00", amountsByLabel["After Fee Percent"])
}

func TestFakeFeesFeesErrorInjection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	injectedErr := assert.AnError

	client := leriantest.NewFakeClient(
		leriantest.WithErrorOn("fees.Fees.Calculate", injectedErr),
	)

	_, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: sampleTransactionDSL(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)
}

func TestFakeFeesFeesValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Nil input
	result, err := client.Fees.Fees.Calculate(ctx, nil)
	require.Error(t, err)
	assert.Nil(t, result)

	// Missing ledger ID
	result, err = client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ledger ID is required")
}

func TestFakeFeesFeesValidationParity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   *fees.FeeCalculate
		message string
	}{
		{name: "missing asset", input: &fees.FeeCalculate{LedgerID: "ledger-1", Transaction: fees.TransactionDSL{Send: fees.TransactionDSLSend{Value: "10.00", Source: fees.TransactionDSLSource{From: []fees.TransactionDSLLeg{{AccountAlias: "sender", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}, Distribute: fees.TransactionDSLDistribute{To: []fees.TransactionDSLLeg{{AccountAlias: "recipient", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}}}}, message: "transaction send asset is required"},
		{name: "missing value", input: &fees.FeeCalculate{LedgerID: "ledger-1", Transaction: fees.TransactionDSL{Send: fees.TransactionDSLSend{Asset: "BRL", Source: fees.TransactionDSLSource{From: []fees.TransactionDSLLeg{{AccountAlias: "sender", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}, Distribute: fees.TransactionDSLDistribute{To: []fees.TransactionDSLLeg{{AccountAlias: "recipient", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}}}}, message: "transaction send value is required"},
		{name: "missing source legs", input: &fees.FeeCalculate{LedgerID: "ledger-1", Transaction: fees.TransactionDSL{Send: fees.TransactionDSLSend{Asset: "BRL", Value: "10.00", Distribute: fees.TransactionDSLDistribute{To: []fees.TransactionDSLLeg{{AccountAlias: "recipient", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}}}}, message: "transaction source legs are required"},
		{name: "missing distribute legs", input: &fees.FeeCalculate{LedgerID: "ledger-1", Transaction: fees.TransactionDSL{Send: fees.TransactionDSLSend{Asset: "BRL", Value: "10.00", Source: fees.TransactionDSLSource{From: []fees.TransactionDSLLeg{{AccountAlias: "sender", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "10.00"}}}}}}}, message: "transaction distribute legs are required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			client := leriantest.NewFakeClient()

			result, err := client.Fees.Fees.Calculate(ctx, tt.input)
			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

func TestFakeFeesFeesReturnsDetachedCopy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	input := &fees.FeeCalculate{
		LedgerID: "ledger-123",
		Transaction: fees.TransactionDSL{
			Metadata: map[string]any{"packageAppliedID": "pkg-1"},
			Send: fees.TransactionDSLSend{
				Asset: "BRL",
				Value: "150.00",
				Source: fees.TransactionDSLSource{
					From: []fees.TransactionDSLLeg{
						{AccountAlias: "@external/BRL", Metadata: map[string]any{"feeLabel": "fee-a"}, Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "150.00"}},
					},
				},
				Distribute: fees.TransactionDSLDistribute{
					To: []fees.TransactionDSLLeg{
						{AccountAlias: "@customer", Amount: &fees.TransactionDSLAmount{Asset: "BRL", Value: "150.00"}},
					},
				},
			},
		},
	}

	output, err := client.Fees.Fees.Calculate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	// Mutate output -- input must remain unchanged
	output.Transaction.Metadata["packageAppliedID"] = "pkg-2"
	output.Transaction.Send.Source.From[0].Metadata["feeLabel"] = "fee-b"

	assert.Equal(t, "pkg-1", input.Transaction.Metadata["packageAppliedID"])
	assert.Equal(t, "fee-a", input.Transaction.Send.Source.From[0].Metadata["feeLabel"])
}

func TestFakeFeesFeesReturnsDetachedCopyForDynamicFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	input := &fees.FeeCalculate{
		LedgerID: "ledger-123",
		Transaction: fees.TransactionDSL{
			TransactionDate: map[string]string{"at": "2026-01-01"},
			Send: fees.TransactionDSLSend{
				Asset: "BRL",
				Value: []string{"150.00"},
				Source: fees.TransactionDSLSource{
					From: []fees.TransactionDSLLeg{
						{
							AccountAlias: "@sender",
							Amount: &fees.TransactionDSLAmount{
								Asset: "BRL",
								Value: []map[string]any{{"gross": "150.00"}},
							},
						},
					},
				},
				Distribute: fees.TransactionDSLDistribute{
					To: []fees.TransactionDSLLeg{
						{
							AccountAlias: "@recipient",
							Amount: &fees.TransactionDSLAmount{
								Asset: "BRL",
								Value: []map[string]any{{"net": "145.00"}},
							},
						},
					},
				},
			},
		},
	}

	output, err := client.Fees.Fees.Calculate(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, output)

	// Mutate output slices/maps -- input must remain unchanged
	outputDate := output.Transaction.TransactionDate.(map[string]string)
	outputDate["at"] = "changed"

	outputValue := output.Transaction.Send.Value.([]string)
	outputValue[0] = "changed"

	outputAmount := output.Transaction.Send.Distribute.To[0].Amount.Value.([]map[string]any)
	outputAmount[0]["net"] = "changed"

	inputDate := input.Transaction.TransactionDate.(map[string]string)
	inputValue := input.Transaction.Send.Value.([]string)
	inputSourceAmount := input.Transaction.Send.Source.From[0].Amount.Value.([]map[string]any)
	inputAmount := input.Transaction.Send.Distribute.To[0].Amount.Value.([]map[string]any)

	assert.Equal(t, "2026-01-01", inputDate["at"])
	assert.Equal(t, "150.00", inputValue[0])
	assert.Equal(t, "150.00", inputSourceAmount[0]["gross"])
	assert.Equal(t, "145.00", inputAmount[0]["net"])
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
			FeeGroupLabel: "Package",
			LedgerID:      "ledger-001",
			MinimumAmount: "0",
			MaximumAmount: "99999",
			Fees:          sampleFeeMap(),
			Enable:        boolPtr(true),
		})
		require.NoError(t, err)
	}

	resp, err := client.Fees.Packages.List(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, resp.Items, 3)
	assert.Equal(t, 3, resp.TotalItems)
}

func TestFakeFeesPackagesListPagination(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create 5 packages
	for i := 0; i < 5; i++ {
		_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
			FeeGroupLabel: "Package",
			LedgerID:      "ledger-001",
			MinimumAmount: "0",
			MaximumAmount: "99999",
			Fees:          sampleFeeMap(),
			Enable:        boolPtr(true),
		})
		require.NoError(t, err)
	}

	// Page 1 with limit 2
	resp, err := client.Fees.Packages.List(ctx, &fees.PackageListOptions{
		PageSize:   2,
		PageNumber: 1,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, 5, resp.TotalItems)
	assert.Equal(t, 1, resp.PageNumber)
	assert.Equal(t, 2, resp.PageSize)

	// Page 3 with limit 2 -- only 1 item left
	resp, err = client.Fees.Packages.List(ctx, &fees.PackageListOptions{
		PageSize:   2,
		PageNumber: 3,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, 5, resp.TotalItems)
}

func TestFakeFeesPackagesListAppliesFilters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	_, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel:    "Retail Enabled",
		SegmentID:        stringPtr("seg-retail"),
		TransactionRoute: stringPtr("ted_out"),
		LedgerID:         "ledger-001",
		MinimumAmount:    "0.00",
		MaximumAmount:    "9999.00",
		Fees:             sampleFeeMap(),
		Enable:           boolPtr(true),
	})
	require.NoError(t, err)

	_, err = client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel:    "Retail Disabled",
		SegmentID:        stringPtr("seg-retail"),
		TransactionRoute: stringPtr("ted_out"),
		LedgerID:         "ledger-001",
		MinimumAmount:    "0.00",
		MaximumAmount:    "9999.00",
		Fees:             sampleFeeMap(),
		Enable:           boolPtr(false),
	})
	require.NoError(t, err)

	_, err = client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel:    "Other Ledger",
		SegmentID:        stringPtr("seg-wholesale"),
		TransactionRoute: stringPtr("pix_out"),
		LedgerID:         "ledger-002",
		MinimumAmount:    "0.00",
		MaximumAmount:    "9999.00",
		Fees:             sampleFeeMap(),
		Enable:           boolPtr(true),
	})
	require.NoError(t, err)

	resp, err := client.Fees.Packages.List(ctx, &fees.PackageListOptions{
		LedgerID:         "ledger-001",
		SegmentID:        "seg-retail",
		TransactionRoute: "ted_out",
		Enabled:          boolPtr(true),
	})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "Retail Enabled", resp.Items[0].FeeGroupLabel)
}

func TestFakeFeesPackagesListValidationParity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()
	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	resp, err := client.Fees.Packages.List(ctx, &fees.PackageListOptions{CreatedFrom: &start})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "createdFrom and createdTo must both be provided")

	resp, err = client.Fees.Packages.List(ctx, &fees.PackageListOptions{CreatedFrom: &start, CreatedTo: &end})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "createdFrom must be before or equal to createdTo")

	resp, err = client.Fees.Packages.List(ctx, &fees.PackageListOptions{SortOrder: "sideways"})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "sortOrder must be either asc or desc")
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
		FeeGroupLabel: "Should Fail",
		LedgerID:      "ledger-001",
		MinimumAmount: "100",
		MaximumAmount: "1000",
		Fees:          sampleFeeMap(),
		Enable:        boolPtr(true),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, injectedErr)

	// Non-injected operations still work
	result, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: sampleTransactionDSL(),
	})
	require.NoError(t, err)
	assert.Equal(t, "ledger-001", result.LedgerID)
}

// ---------------------------------------------------------------------------
// Cross-service: all three fee services work together
// ---------------------------------------------------------------------------

func TestFakeFeesIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	// Create a package
	pkg, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel: "Integration Package",
		LedgerID:      "ledger-001",
		MinimumAmount: "0",
		MaximumAmount: "999999",
		Fees:          sampleFeeMap(),
		Enable:        boolPtr(true),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, pkg.ID)

	txn := sampleTransactionDSL()

	// Use the package ID to calculate an estimate
	estimate, err := client.Fees.Estimates.Calculate(ctx, &fees.FeeEstimateInput{
		PackageID:   pkg.ID,
		LedgerID:    "ledger-001",
		Transaction: txn,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, estimate.Message)
	assert.NotNil(t, estimate.FeesApplied)

	// Use the ledger to calculate an actual fee
	result, err := client.Fees.Fees.Calculate(ctx, &fees.FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: txn,
	})
	require.NoError(t, err)
	assert.Equal(t, "ledger-001", result.LedgerID)

	// Package is still retrievable
	gotPkg, err := client.Fees.Packages.Get(ctx, pkg.ID)
	require.NoError(t, err)
	assert.Equal(t, "Integration Package", gotPkg.FeeGroupLabel)

	// Delete the package
	err = client.Fees.Packages.Delete(ctx, pkg.ID)
	require.NoError(t, err)

	// Package is gone
	_, err = client.Fees.Packages.Get(ctx, pkg.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// Supplemental: optional fields on Create
// ---------------------------------------------------------------------------

func TestFakeFeesPackagesCreateOptionalFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := leriantest.NewFakeClient()

	waived := []string{"acct-vip-001", "acct-vip-002"}
	route := "pix_out"

	created, err := client.Fees.Packages.Create(ctx, &fees.CreatePackageInput{
		FeeGroupLabel:    "Full Options Package",
		Description:      stringPtr("A package with all optional fields"),
		SegmentID:        stringPtr("seg-retail"),
		LedgerID:         "ledger-001",
		TransactionRoute: &route,
		MinimumAmount:    "50",
		MaximumAmount:    "50000",
		WaivedAccounts:   &waived,
		Fees:             sampleFeeMap(),
		Enable:           boolPtr(true),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.NotNil(t, created.Description)
	assert.Equal(t, "A package with all optional fields", *created.Description)
	assert.NotNil(t, created.SegmentID)
	assert.Equal(t, "seg-retail", *created.SegmentID)
	assert.NotNil(t, created.TransactionRoute)
	assert.Equal(t, "pix_out", *created.TransactionRoute)
	assert.NotNil(t, created.WaivedAccounts)
	assert.Len(t, *created.WaivedAccounts, 2)

	// Verify deep clone -- mutating the input slice shouldn't affect stored data
	waived[0] = "mutated"

	got, err := client.Fees.Packages.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "acct-vip-001", (*got.WaivedAccounts)[0])
}
