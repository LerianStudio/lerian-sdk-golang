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

func testTransactionDSL() TransactionDSL {
	return TransactionDSL{
		Description: "TED transfer",
		Route:       "ted_out",
		Pending:     true,
		Metadata: map[string]any{
			"transferType": "TED_OUT",
		},
		Send: TransactionDSLSend{
			Asset: "BRL",
			Value: "15000",
			Source: TransactionDSLSource{
				From: []TransactionDSLLeg{
					{
						Account: "sender",
						Amount:  &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					},
				},
			},
			Distribute: TransactionDSLDistribute{
				To: []TransactionDSLLeg{
					{
						Account: "recipient",
						Amount:  &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					},
				},
			},
		},
	}
}

func transformedTransactionDSL() TransactionDSL {
	tx := testTransactionDSL()
	tx.Send.Value = "15500"
	tx.Send.Source.From = append(tx.Send.Source.From,
		TransactionDSLLeg{
			Account: "sender",
			Amount:  &TransactionDSLAmount{Asset: "BRL", Value: "500"},
			Metadata: map[string]any{
				"feeLabel": "ted_transfer_fee",
			},
		},
	)
	tx.Send.Distribute.To = append(tx.Send.Distribute.To,
		TransactionDSLLeg{
			Account: "platform-fee-account",
			Amount:  &TransactionDSLAmount{Asset: "BRL", Value: "500"},
			Metadata: map[string]any{
				"feeLabel": "ted_transfer_fee",
			},
		},
	)
	tx.Metadata["packageAppliedID"] = "fee-package-uuid"

	return tx
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

func TestFeesTransformTransaction(t *testing.T) {
	t.Parallel()

	expected := transformedTransactionDSL()
	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/fees", path)

			input, ok := body.(*TransformTransactionInput)
			require.True(t, ok)
			assert.Equal(t, "ledger-001", input.LedgerID)
			assert.Equal(t, testTransactionDSL(), input.Transaction)

			return jsonInto(transformResponse{Transaction: &expected}, result)
		},
	}

	svc := newFeesCalcService(backend)
	output, err := svc.TransformTransaction(context.Background(), &TransformTransactionInput{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	})
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, expected, output.Transaction)
}

func TestFeesTransformTransactionNilInput(t *testing.T) {
	t.Parallel()

	svc := newFeesCalcService(&mockBackend{})

	output, err := svc.TransformTransaction(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "Fees.TransformTransaction", sdkErr.Operation)
	assert.Equal(t, "Fee", sdkErr.Resource)
	assert.Contains(t, sdkErr.Message, "input is required")
}

func TestFeesTransformTransactionEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newFeesCalcService(&mockBackend{})

	output, err := svc.TransformTransaction(context.Background(), &TransformTransactionInput{
		Transaction: testTransactionDSL(),
	})
	require.Error(t, err)
	assert.Nil(t, output)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "Fees.TransformTransaction", sdkErr.Operation)
	assert.Equal(t, "Fee", sdkErr.Resource)
	assert.Contains(t, sdkErr.Message, "ledger ID is required")
}

func TestFeesTransformTransactionDataEnvelopeFallback(t *testing.T) {
	t.Parallel()

	expected := transformedTransactionDSL()
	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/fees", path)
			assert.NotNil(t, body)

			return jsonInto(transformResponse{
				Data: &struct {
					Transaction *TransactionDSL `json:"transaction,omitempty"`
				}{Transaction: &expected},
			}, result)
		},
	}

	svc := newFeesCalcService(backend)
	output, err := svc.TransformTransaction(context.Background(), &TransformTransactionInput{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	})
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, expected, output.Transaction)
}

func TestFeesTransformTransactionMissingTransactionInResponse(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/fees", path)
			assert.NotNil(t, body)

			return jsonInto(transformResponse{}, result)
		},
	}

	svc := newFeesCalcService(backend)
	output, err := svc.TransformTransaction(context.Background(), &TransformTransactionInput{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	})
	require.Error(t, err)
	assert.Nil(t, output)
	assert.True(t, errors.Is(err, sdkerrors.ErrInternal))

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "Fees.TransformTransaction", sdkErr.Operation)
	assert.Equal(t, sdkerrors.CategoryInternal, sdkErr.Category)
	assert.Contains(t, sdkErr.Message, "response contained no transaction data")
}

func TestFeesTransformTransactionMissingNestedTransactionInResponse(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/fees", path)
			assert.NotNil(t, body)

			return jsonInto(transformResponse{
				Data: &struct {
					Transaction *TransactionDSL `json:"transaction,omitempty"`
				}{},
			}, result)
		},
	}

	svc := newFeesCalcService(backend)
	output, err := svc.TransformTransaction(context.Background(), &TransformTransactionInput{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	})
	require.Error(t, err)
	assert.Nil(t, output)
	assert.True(t, errors.Is(err, sdkerrors.ErrInternal))

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "Fees.TransformTransaction", sdkErr.Operation)
	assert.Equal(t, sdkerrors.CategoryInternal, sdkErr.Category)
	assert.Contains(t, sdkErr.Message, "response contained no transaction data")
}

func TestFeesTransformTransactionBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("fees backend unavailable")
		},
	}

	svc := newFeesCalcService(backend)
	output, err := svc.TransformTransaction(context.Background(), &TransformTransactionInput{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	})
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "fees backend unavailable")
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ FeesService = (*feesCalcService)(nil)
