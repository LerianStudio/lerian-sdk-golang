package fees

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nilStringer struct{}

func (*nilStringer) String() string { return "10.00" }

// ---------------------------------------------------------------------------
// Shared fixture — builds a realistic TransactionDSL for fee calculation
// ---------------------------------------------------------------------------

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
						AccountAlias: "sender",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					},
				},
			},
			Distribute: TransactionDSLDistribute{
				To: []TransactionDSLLeg{
					{
						AccountAlias: "recipient",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					},
				},
			},
		},
	}
}

var testFeeCalculateResponse = FeeCalculate{
	SegmentID: strPtr("seg-retail"),
	LedgerID:  "ledger-001",
	Transaction: TransactionDSL{
		Description: "TED transfer",
		Route:       "ted_out",
		Pending:     true,
		Metadata: map[string]any{
			"transferType":     "TED_OUT",
			"packageAppliedID": "fee-package-uuid",
		},
		Send: TransactionDSLSend{
			Asset: "BRL",
			Value: "15500",
			Source: TransactionDSLSource{
				From: []TransactionDSLLeg{
					{
						AccountAlias: "sender",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					},
					{
						AccountAlias: "sender",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "500"},
						Metadata: map[string]any{
							"feeLabel": "ted_transfer_fee",
						},
					},
				},
			},
			Distribute: TransactionDSLDistribute{
				To: []TransactionDSLLeg{
					{
						AccountAlias: "recipient",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "15000"},
					},
					{
						AccountAlias: "platform-fee-account",
						Amount:       &TransactionDSLAmount{Asset: "BRL", Value: "500"},
						Metadata: map[string]any{
							"feeLabel": "ted_transfer_fee",
						},
					},
				},
			},
		},
	},
}

// ---------------------------------------------------------------------------
// feesServiceAPI.Calculate — success
// ---------------------------------------------------------------------------

func TestFeesCalculate(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/fees", path)
			assert.NotNil(t, body)

			return jsonInto(testFeeCalculateResponse, result)
		},
	}

	svc := newFeesCalcService(backend)
	input := &FeeCalculate{
		SegmentID:   strPtr("seg-retail"),
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	resp, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "ledger-001", resp.LedgerID)
	assert.NotNil(t, resp.SegmentID)
	assert.Equal(t, "seg-retail", *resp.SegmentID)
	assert.Equal(t, "TED transfer", resp.Transaction.Description)
	assert.Equal(t, "15500", resp.Transaction.Send.Value)

	// The response should have fee legs injected.
	assert.Len(t, resp.Transaction.Send.Source.From, 2, "should have original + fee leg")
	assert.Len(t, resp.Transaction.Send.Distribute.To, 2, "should have original + fee leg")
	assert.Equal(t, "platform-fee-account", resp.Transaction.Send.Distribute.To[1].AccountAlias)
	assert.Equal(t, "ted_transfer_fee", resp.Transaction.Send.Distribute.To[1].Metadata["feeLabel"])
}

// ---------------------------------------------------------------------------
// feesServiceAPI.Calculate — nil input guard clause
// ---------------------------------------------------------------------------

func TestFeesCalculateNilInput(t *testing.T) {
	t.Parallel()

	svc := newFeesCalcService(&mockBackend{})

	resp, err := svc.Calculate(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "Fees.Calculate", sdkErr.Operation)
	assert.Equal(t, "Fee", sdkErr.Resource)
	assert.Contains(t, sdkErr.Message, "input is required")
}

// ---------------------------------------------------------------------------
// feesServiceAPI.Calculate — empty LedgerID guard clause
// ---------------------------------------------------------------------------

func TestFeesCalculateEmptyLedgerID(t *testing.T) {
	t.Parallel()

	svc := newFeesCalcService(&mockBackend{})

	resp, err := svc.Calculate(context.Background(), &FeeCalculate{
		Transaction: testTransactionDSL(),
	})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))

	var sdkErr *sdkerrors.Error
	require.True(t, errors.As(err, &sdkErr))
	assert.Equal(t, "Fees.Calculate", sdkErr.Operation)
	assert.Equal(t, "Fee", sdkErr.Resource)
	assert.Contains(t, sdkErr.Message, "ledger ID is required")
}

func TestFeesCalculateRequiresTransactionShape(t *testing.T) {
	t.Parallel()

	svc := newFeesCalcService(&mockBackend{})

	tests := []struct {
		name    string
		input   *FeeCalculate
		message string
	}{
		{
			name:    "missing asset",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender"}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient"}}}}}},
			message: "transaction send asset is required",
		},
		{
			name:    "missing value",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender"}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient"}}}}}},
			message: "transaction send value is required",
		},
		{
			name:    "blank string value",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "   ", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Share: &TransactionDSLShare{Percentage: 100}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
			message: "transaction send value is required",
		},
		{
			name:    "missing source legs",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient"}}}}}},
			message: "transaction source legs are required",
		},
		{
			name:    "missing distribute legs",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender"}}}}}},
			message: "transaction distribute legs are required",
		},
		{
			name:    "missing leg identifier",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{Amount: &TransactionDSLAmount{Value: "100.00"}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
			message: "transaction source leg identifier is required at index 0",
		},
		{
			name:    "missing amount and share",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender"}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
			message: "transaction source leg amount or share is required at index 0",
		},
		{
			name:    "blank amount value",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Amount: &TransactionDSLAmount{Value: " "}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
			message: "transaction source leg amount value is required at index 0",
		},
		{
			name:    "zero share",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Share: &TransactionDSLShare{Percentage: 0}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient", Share: &TransactionDSLShare{Percentage: 100}}}}}}},
			message: "transaction source leg share must be greater than zero at index 0",
		},
		{
			name:    "distribute leg missing identifier",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Share: &TransactionDSLShare{Percentage: 100}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{Amount: &TransactionDSLAmount{Value: "100.00"}}}}}}},
			message: "transaction distribute leg identifier is required at index 0",
		},
		{
			name:    "distribute leg missing amount and share",
			input:   &FeeCalculate{LedgerID: "ledger-001", Transaction: TransactionDSL{Send: TransactionDSLSend{Asset: "BRL", Value: "100.00", Source: TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender", Share: &TransactionDSLShare{Percentage: 100}}}}, Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient"}}}}}},
			message: "transaction distribute leg amount or share is required at index 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := svc.Calculate(context.Background(), tt.input)
			require.Error(t, err)
			assert.Nil(t, resp)
			assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

func TestFeesCalculateAcceptsBalanceKeyAndPercentageOfPercentage(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			return jsonInto(testFeeCalculateResponse, result)
		},
	}

	svc := newFeesCalcService(backend)
	resp, err := svc.Calculate(context.Background(), &FeeCalculate{
		LedgerID: "ledger-001",
		Transaction: TransactionDSL{Send: TransactionDSLSend{
			Asset: "BRL",
			Value: "100.00",
			Source: TransactionDSLSource{From: []TransactionDSLLeg{{
				BalanceKey: "available",
				Share:      &TransactionDSLShare{PercentageOfPercentage: 100},
			}}},
			Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{
				AccountAlias: "recipient",
				Share:        &TransactionDSLShare{Percentage: 100},
			}}},
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestFeesCalculateRejectsTypedNilSendValue(t *testing.T) {
	t.Parallel()

	svc := newFeesCalcService(&mockBackend{})

	var typedNil *nilStringer

	resp, err := svc.Calculate(context.Background(), &FeeCalculate{
		LedgerID: "ledger-001",
		Transaction: TransactionDSL{Send: TransactionDSLSend{
			Asset:      "BRL",
			Value:      typedNil,
			Source:     TransactionDSLSource{From: []TransactionDSLLeg{{AccountAlias: "sender"}}},
			Distribute: TransactionDSLDistribute{To: []TransactionDSLLeg{{AccountAlias: "recipient"}}},
		}},
	})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "transaction send value is required")
}

// ---------------------------------------------------------------------------
// feesServiceAPI.Calculate — backend error propagation
// ---------------------------------------------------------------------------

func TestFeesCalculateBackendError(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, _ any) error {
			return fmt.Errorf("internal server error")
		},
	}

	svc := newFeesCalcService(backend)
	input := &FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	resp, err := svc.Calculate(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "internal server error")
}

// ---------------------------------------------------------------------------
// feesServiceAPI.Calculate — verifies input body is forwarded
// ---------------------------------------------------------------------------

func TestFeesCalculateInputForwarded(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, body, result any) error {
			input, ok := body.(*FeeCalculate)
			require.True(t, ok, "body should be *FeeCalculate")
			assert.Equal(t, "ledger-002", input.LedgerID)
			assert.NotNil(t, input.SegmentID)
			assert.Equal(t, "seg-premium", *input.SegmentID)
			assert.Equal(t, "BRL", input.Transaction.Send.Asset)

			return jsonInto(testFeeCalculateResponse, result)
		},
	}

	svc := newFeesCalcService(backend)
	input := &FeeCalculate{
		SegmentID:   strPtr("seg-premium"),
		LedgerID:    "ledger-002",
		Transaction: testTransactionDSL(),
	}

	_, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// feesServiceAPI.Calculate — without SegmentID
// ---------------------------------------------------------------------------

func TestFeesCalculateWithoutSegmentID(t *testing.T) {
	t.Parallel()

	resp := FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	backend := &mockBackend{
		callFn: func(_ context.Context, _, _ string, _, result any) error {
			return jsonInto(resp, result)
		},
	}

	svc := newFeesCalcService(backend)
	input := &FeeCalculate{
		LedgerID:    "ledger-001",
		Transaction: testTransactionDSL(),
	}

	got, err := svc.Calculate(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.SegmentID)
	assert.Equal(t, "ledger-001", got.LedgerID)
}

// ---------------------------------------------------------------------------
// Compile-time interface assertion
// ---------------------------------------------------------------------------

var _ feesServiceAPI = (*feesCalcService)(nil)
