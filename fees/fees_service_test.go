package fees

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Shared fixture
// ---------------------------------------------------------------------------

var testFee = Fee{
	ID:            "fee-001",
	PackageID:     "pkg-001",
	TransactionID: strPtr("txn-001"),
	Amount:        10000,
	Scale:         2,
	Currency:      "USD",
	FeeResults: []FeeResult{
		{RuleType: "flat", Amount: 100, Scale: 2, Currency: "USD", Applied: true},
		{RuleType: "percentage", Amount: 250, Scale: 2, Currency: "USD", Applied: true},
	},
	TotalFee:      350,
	TotalFeeScale: 2,
	Status:        "calculated",
	CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
}

// ---------------------------------------------------------------------------
// FeesService.Calculate — success
// ---------------------------------------------------------------------------

func TestFeesCalculate(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/fees/calculate", path)
			assert.NotNil(t, body)

			return jsonInto(testFee, result)
		},
	}

	svc := newFeesCalcService(backend)
	txnID := "txn-001"
	input := &CalculateFeeInput{
		PackageID:     "pkg-001",
		TransactionID: &txnID,
		Amount:        10000,
		Scale:         2,
		Currency:      "USD",
	}

	fee, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, fee)
	assert.Equal(t, "fee-001", fee.ID)
	assert.Equal(t, "pkg-001", fee.PackageID)
	assert.Equal(t, "txn-001", *fee.TransactionID)
	assert.Equal(t, int64(10000), fee.Amount)
	assert.Equal(t, "USD", fee.Currency)
	assert.Len(t, fee.FeeResults, 2)
	assert.Equal(t, int64(350), fee.TotalFee)
	assert.Equal(t, "calculated", fee.Status)
}

// ---------------------------------------------------------------------------
// FeesService.Calculate — nil input guard clause
// ---------------------------------------------------------------------------

func TestFeesCalculateNilInput(t *testing.T) {
	svc := newFeesCalcService(&mockBackend{})

	fee, err := svc.Calculate(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, fee)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	// Verify the error has the right operation and resource.
	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "Fees.Calculate", sdkErr.Operation)
	assert.Equal(t, "Fee", sdkErr.Resource)
	assert.Contains(t, sdkErr.Message, "input is required")
}

// ---------------------------------------------------------------------------
// FeesService.Calculate — backend error propagation
// ---------------------------------------------------------------------------

func TestFeesCalculateBackendError(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("internal server error")
		},
	}

	svc := newFeesCalcService(backend)
	input := &CalculateFeeInput{
		PackageID: "pkg-001",
		Amount:    10000,
		Scale:     2,
		Currency:  "USD",
	}

	fee, err := svc.Calculate(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, fee)
	assert.Contains(t, err.Error(), "internal server error")
}

// ---------------------------------------------------------------------------
// FeesService.Calculate — verifies input body is forwarded
// ---------------------------------------------------------------------------

func TestFeesCalculateInputForwarded(t *testing.T) {
	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, body, result any) error {
			input, ok := body.(*CalculateFeeInput)
			require.True(t, ok, "body should be *CalculateFeeInput")
			assert.Equal(t, "pkg-003", input.PackageID)
			assert.Equal(t, int64(25000), input.Amount)
			assert.Equal(t, "BRL", input.Currency)
			assert.Nil(t, input.TransactionID)

			return jsonInto(testFee, result)
		},
	}

	svc := newFeesCalcService(backend)
	input := &CalculateFeeInput{
		PackageID: "pkg-003",
		Amount:    25000,
		Scale:     2,
		Currency:  "BRL",
	}

	_, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// FeesService.Calculate — without TransactionID
// ---------------------------------------------------------------------------

func TestFeesCalculateWithoutTransactionID(t *testing.T) {
	feeNoTxn := Fee{
		ID:            "fee-002",
		PackageID:     "pkg-001",
		TransactionID: nil,
		Amount:        5000,
		Scale:         2,
		Currency:      "USD",
		FeeResults:    []FeeResult{},
		TotalFee:      50,
		TotalFeeScale: 2,
		Status:        "calculated",
		CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			return jsonInto(feeNoTxn, result)
		},
	}

	svc := newFeesCalcService(backend)
	input := &CalculateFeeInput{
		PackageID: "pkg-001",
		Amount:    5000,
		Scale:     2,
		Currency:  "USD",
	}

	fee, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, fee)
	assert.Nil(t, fee.TransactionID)
	assert.Equal(t, "fee-002", fee.ID)
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ FeesService = (*feesCalcService)(nil)
